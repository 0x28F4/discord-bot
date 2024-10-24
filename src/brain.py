from typing import Callable, Iterable
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types
from chat import Chat, ChatMessage
from tts import tts
from utils import DEBUG


with open(".prompt", "r") as pf:
    SYSTEM_PROMPT = pf.read()


def echo(
    responses: Iterable[cloud_speech_types.StreamingRecognizeResponse],
    on_audio: Callable[[bytes], None],
) -> None:
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


chat = Chat(llm_name="Jane", system_prompt=SYSTEM_PROMPT)


def respond(
    *,
    user: str,
    responses: Iterable[cloud_speech_types.StreamingRecognizeResponse],
    on_audio: Callable[[bytes], None],
) -> None:
    for response in responses:
        if not response.results:
            continue
        result = response.results[0]

        if not result.alternatives:
            continue
        transcript = result.alternatives[0].transcript

        if DEBUG():
            print(f"final={result.is_final}|transcript={transcript}")
        if not result.is_final:
            continue
        chat.say(ChatMessage(user_name=user, content=transcript))
        if not (response := chat.complete()):
            continue
        on_audio(tts(response))
