from typing import Callable, Iterable, Optional
from chat import Chat, ChatMessage
from discord_audio import AudioData
from stt import STT
from utils import DEBUG
from tts import TTS

class Engine():
    def __init__(self, *, chat: Chat, tts: TTS, on_audio: Callable[[bytes], None]):
        self.chat = chat
        self.tts = tts
        self.stt = STT()
        self.on_audio = on_audio
    
    def run(self, audio_stream: Iterable[AudioData], user: str):
        while True:
            last_transcription: Optional[str] = None
            for response in self.stt.stream(source=audio_stream):
                if not response.results:
                    continue
                result = response.results[0]

                if not result.alternatives:
                    continue
                transcript = result.alternatives[0].transcript

                if DEBUG(): print(f"final={result.is_final}|transcript={transcript}")
                if not result.is_final:
                    last_transcription = transcript
                    continue
                last_transcription = None
                self.chat.say(ChatMessage(user_name=user, content=transcript))
                if not (response := self.chat.complete()):
                    continue
                self.on_audio(self.tts.convert(response))

            if DEBUG(): print("reached end of speech request.")
            if last_transcription:
                self.chat.say(ChatMessage(user_name=user, content=transcript))
                if not (response := self.chat.complete()):
                    continue
                self.on_audio(self.tts.convert(response))