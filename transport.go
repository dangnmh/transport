package transport

import (
	"net/http"
)

func New(tp *http.Transport, opts ...TransportOption) http.RoundTripper {
	if tp == nil {
		tp = http.DefaultTransport.(*http.Transport).Clone()
	}

	for _, opts := range opts {
		opts(tp)
	}

	return tp
}
