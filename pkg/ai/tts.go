package ai

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

var client = &http.Client{}

func tts(text, filepath string) error {
	path := fmt.Sprintf("http://localhost:5002/api/tts?text=%s&speaker_id=p364&style_wav=&language_id=", url.QueryEscape(text))
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
