from pathlib import Path
import re
from typing import TypedDict
import yaml

class TTSConfig(TypedDict):
    language_code: str
    voice: str

class ChatConfig(TypedDict):
    name: str
    host: str
    model: str
    bot_name: str
    system_prompt: str

class Config(TypedDict):
    chat: ChatConfig
    tts: TTSConfig

def load(name: str = "example") -> Config:
    file_path = Path(__file__).parent.resolve() / f"config.{name}.yaml"
    with file_path.open('r') as f:
        return yaml.safe_load(f.read())


def list():
    def _parse(file_name: str):
        matches = re.search(r"^config\.(.*)\.yaml$", file_name)
        if not matches:
            return
        return matches.group(1)

    return [
        parsed
        for p in Path(__file__).parent.resolve().iterdir()
        if (parsed := _parse(p.name))
    ]
