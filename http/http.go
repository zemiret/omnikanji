package http

import (
	"fmt"
	"net/http"
)

func StatusCodeError(code int) error {
	return fmt.Errorf("status code: %d", code)
}

func Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, StatusCodeError(resp.StatusCode)
	}

	return resp, nil
}

