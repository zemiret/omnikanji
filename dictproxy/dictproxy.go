package dictproxy

import "net/http"

type HttpClient interface {
	Get(url string) (*http.Response, error) 
}