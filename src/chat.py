import re
from typing import List

from openai import OpenAI
from retry import retry
from config import ChatConfig
import config
from utils import DEBUG

class ChatMessage:
    def __init__(self, user_name: str, content: str):
        self.user_name = user_name
        self.content = content

    def __str__(self):
        return f"[{self.user_name}]: {self.content}"

class Chat:
    def __init__(self, config: ChatConfig):
        self.openai_client = OpenAI(api_key="local", base_url=config["host"])
        self.bot_name = config["bot_name"]
        self.model = config["model"]
        self.history: List[ChatMessage | str] = [config["system_prompt"]]

    def _extract_content(self, text: str):
        pattern = r'\[.*?\]'
        match = re.search(pattern, text)
        if match:
            return text[:match.start()].strip()
        else:
            return text.strip()

    def _format_history(self):
        messages = [f"{msg}" for msg in self.history]
        return "\n".join(messages)

    def say(self, message: ChatMessage):
        assert isinstance(message, ChatMessage)
        self.history.append(message)

    @retry(tries=3, delay=0)
    def complete(self):
        self.history.append(ChatMessage(user_name=self.bot_name, content=""))
        completions = self.openai_client.chat.completions.create(
            max_tokens=100,
            messages=[{"role": "user", "content": self._format_history()}],
            model=self.model,
        )

        if extracted_result := self._extract_content(
            completions.choices[0].message.content
        ):
            self.history = self.history[0:-1]
            self.history.append(ChatMessage(user_name=self.bot_name, content=extracted_result))
            if DEBUG(): print("complete", self._format_history())
            return extracted_result
        self.history = self.history[0:-1]

if __name__ == "__main__":
    chat = Chat(config=config.load()["chat"])
    while True:
        print(">", end="")
        chat.say(ChatMessage(user_name="hero", content=input()))
        chat.complete()
