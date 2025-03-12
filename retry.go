package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/cenkalti/backoff/v4"
)

type retryTransport struct {
	tp     http.RoundTripper
	config *retryConfig
}

type retryConfig struct {
	RetryOnError   bool
	OnStatus       []int
	MaxTries       uint64   // Maximum number of retry attempts.
	WhiteListPaths []string // * mean all example GET|/v1/user/get
	BlackListPaths []string
}

var DefaultRetryConfig = retryConfig{
	RetryOnError:   true,
	OnStatus:       []int{429, 502, 503, 504},
	MaxTries:       10,
	WhiteListPaths: []string{ConsCharStar},
}

func NewTransportRetry(tp http.RoundTripper, opts ...RetryOption) http.RoundTripper {
	cfg := DefaultRetryConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return &retryTransport{
		tp:     tp,
		config: &cfg,
	}
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloneReq, err := cloneRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := rt.tp.RoundTrip(cloneReq)
	if !rt.validRetryPath(req) {
		return res, err
	}

	if err != nil && !rt.config.RetryOnError {
		return res, err
	}

	if res != nil && !slices.Contains(rt.config.OnStatus, res.StatusCode) {
		return res, err
	}
	fmt.Println(rt.config.MaxTries,"222222222")

	bo := backoff.NewExponentialBackOff()

	res, err = backoff.RetryWithData(func() (*http.Response, error) {
		cloneReq, err := cloneRequest(req)
		if err != nil {
			return nil, err
		}

		return rt.tp.RoundTrip(cloneReq)
	}, backoff.WithMaxRetries(bo, rt.config.MaxTries))

	return res, err
}

func (rt *retryTransport) validRetryPath(req *http.Request) bool {
	if slices.Contains(rt.config.BlackListPaths, ConsCharStar) {
		return false
	}

	method := req.Method
	path := req.URL.Path
	combine := fmt.Sprintf("%s%s%s", method, ConsCharVerticalBar, path)

	if slices.Contains(rt.config.BlackListPaths, combine) {
		return false
	}

	if slices.Contains(rt.config.WhiteListPaths, ConsCharStar) {
		return true
	}

	if slices.Contains(rt.config.WhiteListPaths, combine) {
		return true
	}

	return false
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	newReq := req.Clone(req.Context()) // Clone method copies headers and context

	// If the request has a body, we need to copy it
	if req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, req.Body) // Copy body content
		if err != nil {
			return nil, err
		}

		newReq.Body = io.NopCloser(bytes.NewReader(buf.Bytes())) // Reset body
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))    // Reset original body
	}

	return newReq, nil
}
