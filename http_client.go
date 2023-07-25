package tus

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
