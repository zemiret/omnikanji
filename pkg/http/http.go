package http

import (
	"fmt"
	"net/http"
)

type Client struct {}

func statusCodeError(code int) error {
	return fmt.Errorf("status code: %d", code)
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return resp, statusCodeError(resp.StatusCode)
	}

	return resp, nil
}

