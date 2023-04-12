package tts

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type coqui struct {
	client  *http.Client
	address string
	voice   string
}

func NewCoqui(address, voice string) TTS {
	if voice == "" {
		voice = "p364"
	}
	return &coqui{
		client:  &http.Client{},
		address: address,
		voice:   voice,
	}
}

func (t *coqui) ToFile(text, filepath string) (string, error) {
	path := fmt.Sprintf("%s/api/tts?text=%s&speaker_id=%s&style_wav=&language_id=", t.address, url.QueryEscape(text), t.voice)
	fmt.Printf("sending request to %s\n", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		resultPath := filepath + ".wav"
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

func (t *coqui) ChangeVoice(v string) {
	t.voice = v
}
