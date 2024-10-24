
from brain import Chat, ChatMessage

def test_chat_message():
    msg = ChatMessage("foo", "bar")
    assert f"{msg}" == "[foo]: bar"

def test_chat():
    chat = Chat(llm_name="robot", system_prompt="custom system prompt")
    assert len(chat.history) == 1
    assert chat.history[0] == "custom system prompt"
    chat.say(ChatMessage("foo", "content"))
    assert len(chat.history) == 2
    assert chat._format_history() == "custom system prompt\n[foo]: content"
