import os
from typing import Callable, Iterable, List, TypedDict
from google.cloud.speech_v2.types import cloud_speech as cloud_speech_types
from openai import OpenAI
import re
from tts import tts
from utils import DEBUG


with open('.prompt', 'r') as pf:
    SYSTEM_PROMPT = pf.read()

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

class ChatMessage():
    def __init__(self, user_name: str, content: str):
        self.user_name = user_name
        self.content = content

    def __str__(self):
        return f"[{self.user_name}]: {self.content}"

class Chat():
    def __init__(self, llm_name: str, system_prompt: str, llm_host: str = "http://localhost:11434/v1"):
        self.openai_client = OpenAI(
            api_key="local",
            base_url=llm_host
        )
        self.llm_name = llm_name
        self.history: List[ChatMessage] = [system_prompt]

    def _extract_content(self, text: str):
        pattern  = fr'\[{self.llm_name}\]:(.*)'
        match = re.search(pattern, text)
        if match:
            return match.group(1).strip()
        return None
    
    def _format_history(self):
        messages = [f"{msg}" for msg in self.history]
        return '\n'.join(messages)

    def say(self, message: ChatMessage):
        self.history.append(message)

    def complete(self):
        completions = self.openai_client.chat.completions.create(
            max_tokens=100,
            messages=[
                {
                    "role": "user",
                    "content": self._format_history()
                }
            ],
            model="hf.co/TheBloke/opus-v0-7B-GGUF:Q6_K",
        )

        if (extracted_result := self._extract_content(completions.choices[0].message.content)):
            self.history.append(ChatMessage(user_name=self.llm_name, content=extracted_result))
            if DEBUG(): print("complete", self._format_history())
            return extracted_result

chat = Chat(llm_name="Jane", system_prompt=SYSTEM_PROMPT)

def respond(*, user: str, responses: Iterable[cloud_speech_types.StreamingRecognizeResponse], on_audio: Callable[[bytes], None]) -> None:
    for response in responses:
        if not response.results: continue
        result = response.results[0]

        if not result.alternatives: continue
        transcript = result.alternatives[0].transcript
        
        if DEBUG(): print(f"final={result.is_final}|transcript={transcript}")
        if not result.is_final: continue
        chat.say(ChatMessage(user_name=user, content=transcript))
        if not (response := chat.complete()): continue
        on_audio(tts(response))
