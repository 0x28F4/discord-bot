from engine import Chat, ChatMessage
from config import ChatConfig


def test_chat_message():
    msg = ChatMessage("foo", "bar")
    assert f"{msg}" == "[foo]: bar"


config = ChatConfig(
    name="foo",
    host="localhost",
    model="gpt-3.5",
    bot_name="foo",
    system_prompt="custom system prompt",
)


def test_chat():
    chat = Chat(config=config)
    assert len(chat.history) == 1
    assert chat.history[0] == "custom system prompt"
    chat.say(ChatMessage("foo", "content"))
    assert len(chat.history) == 2
    assert chat._format_history() == "custom system prompt\n[foo]: content"
