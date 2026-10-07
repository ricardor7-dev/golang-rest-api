package server

import(
	"net/http"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
} 


var (
	Client HTTPClient
)

func init() {
	Client = &http.Client{
		Timeout: time.Second * 20,
	}
}

