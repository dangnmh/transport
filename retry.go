package transport

import (
	"errors"
	"fmt"
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
	if !rt.validRetryPath(req) || rt.config.MaxTries == 0 {
		return res, err
	}

	if err != nil && !rt.config.RetryOnError {
		return res, err
	}

	if !rt.validRetryResStatus(res) {
		return res, err
	}

	bo := backoff.NewExponentialBackOff()

	var lastSuccessRes *http.Response
	res, err = backoff.RetryWithData(func() (*http.Response, error) {
		cloneReq, err := cloneRequest(req)
		if err != nil {
			return nil, err
		}

		res, err := rt.tp.RoundTrip(cloneReq)
		if err != nil {
			return nil, err
		}

		lastSuccessRes = res
		if rt.validRetryResStatus(res) {
			return nil, errors.New("bad status")
		}

		return res, err
	}, backoff.WithMaxRetries(bo, rt.config.MaxTries))
	if err != nil && lastSuccessRes != nil {
		return lastSuccessRes, nil
	}

	return res, err
}

func (rt *retryTransport) validRetryResStatus(res *http.Response) bool {
	if res != nil && !slices.Contains(rt.config.OnStatus, res.StatusCode) {
		return false
	}

	return true
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
