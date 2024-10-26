from typing import Callable, Iterable
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types
from chat import Chat, ChatMessage
from utils import DEBUG
from tts import TTS

class Engine():
    def __init__(self, *, chat: Chat, tts: TTS, on_audio: Callable[[bytes], None]):
        self.chat = chat
        self.tts = tts
        self.on_audio = on_audio
    
    def run(self, responses: Iterable[cloud_speech_types.StreamingRecognizeResponse], user: str):
        for response in responses:
            if not response.results:
                continue
            result = response.results[0]

            if not result.alternatives:
                continue
            transcript = result.alternatives[0].transcript

            if DEBUG(): print(f"final={result.is_final}|transcript={transcript}")
            if not result.is_final:
                continue
            self.chat.say(ChatMessage(user_name=user, content=transcript))
            if not (response := self.chat.complete()):
                continue
            self.on_audio(self.tts.convert(response))
