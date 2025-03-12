package transport

import (
	"bytes"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type retryTransport struct {
	tp     http.RoundTripper
	config *retryConfig
}

type retryConfig struct {
	OnStatus       []int
	MaxTries       uint64        // Maximum number of retry attempts.
	MaxElapsedTime time.Duration // Maximum total time for all retries.
}

var DefaultRetryConfig = &retryConfig{
	OnStatus:       []int{429, 502, 503, 504},
	MaxTries:       10,
	MaxElapsedTime: 15 * time.Minute,
}

func NewTransportRetry(tp http.RoundTripper, opts ...RetryOption) http.RoundTripper {
	cfg := &*DefaultRetryConfig
	for _, opt := range opts {
		opt(cfg)
	}

	return &retryTransport{
		tp:     tp,
		config: cfg,
	}
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloneReq, err := cloneRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := rt.tp.RoundTrip(cloneReq)
	if err != nil || !rt.isRetryable(res.StatusCode) {
		return res, err
	}

	bo := backoff.NewExponentialBackOff(backoff.WithMaxElapsedTime(rt.config.MaxElapsedTime))

	// var lastRes *http.Response
	res, err = backoff.RetryWithData(func() (*http.Response, error) {
		cloneReq, err := cloneRequest(req)
		if err != nil {
			return nil, err
		}

		return rt.tp.RoundTrip(cloneReq)
	}, backoff.WithMaxRetries(bo, rt.config.MaxTries))

	return res, err
}

func (rt *retryTransport) isRetryable(status int) bool {
	return slices.Contains(rt.config.OnStatus, status)
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
