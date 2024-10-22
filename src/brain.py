from typing import Callable, Iterable
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types

from tts import tts

def echo(responses: Iterable[cloud_speech_types.StreamingRecognizeResponse], on_audio: Callable[[bytes], None]) -> None:
    """
    echos everything thats said back to the user
    """
    for response in responses:
        if not response.results:
            continue

        result = response.results[0]

        if not result.alternatives:
            continue

        transcript = result.alternatives[0].transcript

        if result.is_final:
            on_audio(tts(transcript))
