import re
from typing import List

from openai import OpenAI
from utils import DEBUG


class ChatMessage:
    def __init__(self, user_name: str, content: str):
        self.user_name = user_name
        self.content = content

    def __str__(self):
        return f"[{self.user_name}]: {self.content}"


class Chat:
    def __init__(
        self,
        llm_name: str,
        system_prompt: str,
        llm_host: str = "http://localhost:11434/v1",
    ):
        self.openai_client = OpenAI(api_key="local", base_url=llm_host)
        self.llm_name = llm_name
        self.history: List[ChatMessage] = [system_prompt]

    def _extract_content(self, text: str):
        pattern = rf"\[{self.llm_name}\]:(.*)"
        match = re.search(pattern, text)
        if match:
            return match.group(1).strip()
        return None

    def _format_history(self):
        messages = [f"{msg}" for msg in self.history]
        return "\n".join(messages)

    def say(self, message: ChatMessage):
        self.history.append(message)

    def complete(self):
        completions = self.openai_client.chat.completions.create(
            max_tokens=100,
            messages=[{"role": "user", "content": self._format_history()}],
            model="hf.co/TheBloke/opus-v0-7B-GGUF:Q6_K",
        )

        if extracted_result := self._extract_content(
            completions.choices[0].message.content
        ):
            self.history.append(
                ChatMessage(user_name=self.llm_name, content=extracted_result)
            )
            if DEBUG():
                print("complete", self._format_history())
            return extracted_result
