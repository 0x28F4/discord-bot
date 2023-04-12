package tts

type TTS interface {
	ToFile(text, filepath string) (string, error)
	ChangeVoice(v string)
}
