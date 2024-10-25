from pathlib import Path
import re
from typing import TypedDict
import yaml


class Config(TypedDict):
    name: str
    host: str
    model: str
    bot_name: str
    system_prompt: str


def load(name: str = "example"):
    with open(f"config.{name}.yaml", "r") as f:
        yaml.safe_load(f.read())


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
