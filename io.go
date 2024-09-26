package filegunner

import (
	"io"
	"net/http"
)

type CreateFileFunc func(string) (io.WriteCloser, error)
type ReadFileFunc func(string) ([]byte, error)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HttpClientWrapper struct {
	client *http.Client
}

func NewHttpClientWrapper() *HttpClientWrapper {
	return &HttpClientWrapper{
		client: &http.Client{},
	}
}

func (h *HttpClientWrapper) Do(req *http.Request) (*http.Response, error) {
	return h.client.Do(req)
}
