package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var client *http.Client

func init() {
	client = &http.Client{}
}

func makeBody(payload any) (io.Reader, error) {
	if payload == nil {
		return nil, nil
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return nil, err
	}

	return &body, nil
}

func Request[T any](method, path string, headers map[string]string, payload any) (T, error) {
	var result T
	body, err := makeBody(payload)
	if err != nil {
		return result, err
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return result, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return result, err
		}

		return result, nil
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	return result, fmt.Errorf("unexpected status code: %d, err: %s", resp.StatusCode, string(bts))
}
