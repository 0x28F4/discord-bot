from google.cloud import texttospeech

from config import TTSConfig


client = texttospeech.TextToSpeechClient()

class TTS():
    def __init__(self, config: TTSConfig):
        self.voice = texttospeech.VoiceSelectionParams(
            language_code=config["language_code"],
            name=config["voice"],
            ssml_gender=texttospeech.SsmlVoiceGender.FEMALE,
        )
        self.audio_config = texttospeech.AudioConfig(
            audio_encoding=texttospeech.AudioEncoding.LINEAR16,
            sample_rate_hertz=48000,
        )

    def convert(self, content: str):
        synthesis_input = texttospeech.SynthesisInput(text=content)
        response = client.synthesize_speech(
            input=synthesis_input, voice=self.voice, audio_config=self.audio_config
        )
        return response.audio_content

