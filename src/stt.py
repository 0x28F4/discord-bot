import os
from typing import Any, Generator, Iterable, List
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types
from google.cloud.speech_v2 import SpeechClient

from discord_audio import AudioData

SPEECH_MAX_CHUNK_SIZE = 25600

class STT():
    def __init__(self):
        assert (GCP_PROJECT_ID := os.getenv("DISCORD_BOT_GCP_ID"))
        self.client = SpeechClient()
        recognition_config = cloud_speech_types.RecognitionConfig(
            explicit_decoding_config=cloud_speech_types.ExplicitDecodingConfig(
                sample_rate_hertz=48000,
                encoding=cloud_speech_types.ExplicitDecodingConfig.AudioEncoding.LINEAR16,
                audio_channel_count=2,
            ),
            language_codes=["en-US"],
            model="long",
        )
        streaming_config = cloud_speech_types.StreamingRecognitionConfig(
            config=recognition_config,
            streaming_features=cloud_speech_types.StreamingRecognitionFeatures(
                interim_results=True
            ),
        )
        self.config_request = cloud_speech_types.StreamingRecognizeRequest(
            recognizer=f"projects/{GCP_PROJECT_ID}/locations/global/recognizers/_",
            streaming_config=streaming_config,
        )

    def stream(self, *, source: Generator[bytes, Any, None]) -> Iterable[cloud_speech_types.StreamingRecognizeResponse]:
        def _requests_gen(config: cloud_speech_types.RecognitionConfig, audio_data: List[AudioData]):
            started = False
            for audio in audio_data:
                if audio.silence:
                    return
                buffer = audio.data
                if not started:
                    yield config
                    started = True
                for i in range(0, len(buffer), SPEECH_MAX_CHUNK_SIZE):
                    chunk = buffer[i:SPEECH_MAX_CHUNK_SIZE]
                    # must turn off discords voice activation, otherwise this won't recognize silence
                    if any(chunk):
                        yield cloud_speech_types.StreamingRecognizeRequest(audio=chunk)
        return self.client.streaming_recognize(
            requests=_requests_gen(self.config_request, source)
        )
