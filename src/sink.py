import os
from typing import List
import discord as dc

from google.cloud.speech_v2 import SpeechClient
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types

class RecognizeSink(dc.sinks.Sink):
    def __init__(self, *, filters=None):
        if filters is None:
            filters = dc.sinks.default_filters
        self.filters = filters
        dc.sinks.Filters.__init__(self, **self.filters)
        self.vc = None
        self.audio_data = {}

        # user id for parsing their specific audio data
        self.user_id = None

        # obj to store our super sweet awesome audio data
        self.buffer = Transcriber()

    def write(self, data, user):
        # we overload the write method to take advantage of the already running thread for recording

        # if the data comes from the inviting user, we append it to buffer
        if user == self.user_id:
            self.buffer.write(data=data, user=user)

    def cleanup(self):
        self.finished = True

    def get_all_audio(self):
        # not applicable for streaming but may cause errors if not overloaded
        pass

    def get_user_audio(self, user):
        # not applicable for streaming but will def cause errors if not overloaded called
        pass

    def set_user(self, user_id: int):
        self.user_id = user_id
        print(f"Set user ID: {user_id}")


class Transcriber:
    def __init__(self) -> None:
        assert (gcp_id := os.getenv("DISCORD_BOT_GCP_ID"))
        self.gcp_id = gcp_id
        self.client = SpeechClient()

        # holds byte-form audio data as it builds
        self.byte_buffer = bytearray()  # bytes

        # audio data specifications
        self.sample_width = 2
        self.channels = 2
        self.sample_rate = 48000
        self.bytes_ps = 192000  # bytes added to buffer per second
        self.block_len = 1
        self.buff_lim = int(self.bytes_ps * self.block_len)


    # will need 'user' param if tracking multiple peoples voices - TBD
    def write(self, data, user) -> None:
        self.byte_buffer += data  # data is a bytearray object
        if len(self.byte_buffer) > self.buff_lim:
            segment = self.byte_buffer[:self.buff_lim]
            print(f"got new segment: {len(segment)}")
            self.transcribe(segment)
            self.byte_buffer = self.byte_buffer[self.buff_lim:]


    def transcribe(self, audio: bytearray) -> cloud_speech_types.StreamingRecognizeResponse:
        chunk_size = 25600
        audio_requests = [
            cloud_speech_types.StreamingRecognizeRequest(audio=bytes(audio[i:i+chunk_size])) for i in range(0, len(audio), chunk_size)
        ]

        recognition_config = cloud_speech_types.RecognitionConfig(
            # auto_decoding_config=cloud_speech_types.AutoDetectDecodingConfig(),
            explicit_decoding_config=cloud_speech_types.ExplicitDecodingConfig(
                encoding="LINEAR16",
                sample_rate_hertz=48000,
                audio_channel_count=2,
            ),
            language_codes=["en-US"],
            model="long",
        )
        streaming_config = cloud_speech_types.StreamingRecognitionConfig(
            config=recognition_config
        )
        config_request = cloud_speech_types.StreamingRecognizeRequest(
            recognizer=f"projects/{self.gcp_id}/locations/global/recognizers/_",
            streaming_config=streaming_config,
        )

        def requests(config: cloud_speech_types.RecognitionConfig, audio: list):
            yield config
            yield from audio

        # Transcribes the audio into text
        responses_iterator = self.client.streaming_recognize(
            requests=requests(config_request, audio_requests)
        )
        responses = []
        for response in responses_iterator:
            responses.append(response)
            for result in response.results:
                if len(result.alternatives) == 0:
                    print("no alternatives", response)
                    continue
                print("alternatives", response)
                print(f"Transcript: {result.alternatives[0].transcript}")

        return responses
