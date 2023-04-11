package ai

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type tts struct {
	address string
	voice   string
}

func NewTTS(address string) *tts {
	return &tts{
		address: address,
		voice:   "p364",
	}
}

var client = &http.Client{}

func (t *tts) ToFile(text, filepath string) error {
	path := fmt.Sprintf("%s/api/tts?text=%s&speaker_id=%s&style_wav=&language_id=", t.address, url.QueryEscape(text), t.voice)
	fmt.Printf("sending request to %s\n", path)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		f, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return err
		}
		return nil
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("unexpected status code: %d, err: %s", resp.StatusCode, string(bts))
}

func (t *tts) ChangeVoice(v string) {
	t.voice = v
}
