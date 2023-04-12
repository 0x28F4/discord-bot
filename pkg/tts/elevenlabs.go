package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type elevenlabs struct {
	apiKey string
	voice  string
}

func NewElevenlabs(apiKey string, voice string) TTS {
	if voice == "" {
		voice = "21m00Tcm4TlvDq8ikWAM"
	}
	return &elevenlabs{
		apiKey: apiKey,
		voice:  voice,
	}
}

var client = &http.Client{}

type elevenlabsVoiceSettings struct {
	Stability       float32 `json:"stability"`
	SimilarityBoost float32 `json:"similarity_boost"`
}

type elevenlabsPayload struct {
	Text          string                  `json:"text"`
	VoiceSettings elevenlabsVoiceSettings `json:"voice_settings"`
}

func (t *elevenlabs) ToFile(text, filepath string) (string, error) {
	p := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", t.voice)
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(elevenlabsPayload{
		Text:          text,
		VoiceSettings: elevenlabsVoiceSettings{0.75, 0.75},
	}); err != nil {
		return "", err
	}
	fmt.Printf("sending request to %s\n", p)
	req, err := http.NewRequest(http.MethodPost, p, &body)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		resultPath := filepath + ".mpeg"
		f, err := os.Create(resultPath)
		if err != nil {
			return "", err
		}
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return "", err
		}
		return resultPath, nil
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf("unexpected status code: %d, err: %s", resp.StatusCode, string(bts))
}

func (t *elevenlabs) ChangeVoice(v string) {
	t.voice = v
}
