package templateClient

import "net/http"

type TemplateClient interface {
	Do(req *http.Request) (*http.Response, error)
}
