package transport

import (
	"errors"
	"net/http"

	"github.com/cenkalti/backoff/v4"
)

type retryTransport struct {
	tp      http.RoundTripper
	config  *retryConfig
	matcher Matcher
}

type retryConfig struct {
	RetryOnError bool
	MaxTries     uint64 // Maximum number of retry attempts.
	*MatcherConfig
}

var DefaultRetryConfig = retryConfig{
	RetryOnError:  true,
	MaxTries:      10,
	MatcherConfig: &DefaultMatcherConfig,
}

func NewTransportRetry(tp http.RoundTripper, opts ...RetryOption) http.RoundTripper {
	cfg := DefaultRetryConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return &retryTransport{
		tp:      tp,
		config:  &cfg,
		matcher: NewMatcher(*cfg.MatcherConfig),
	}
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloneReq, err := cloneRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := rt.tp.RoundTrip(cloneReq)
	if rt.config.MaxTries == 0 {
		return res, err
	}

	if err != nil && !rt.config.RetryOnError {
		return res, err
	}

	if res != nil && !rt.matcher.ShouldDo(req, res.StatusCode) {
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
		if rt.matcher.ShouldDo(req, res.StatusCode) {
			return nil, errors.New("bad status")
		}

		return res, err
	}, backoff.WithMaxRetries(bo, rt.config.MaxTries))
	if err != nil && lastSuccessRes != nil {
		return lastSuccessRes, nil
	}

	return res, err
}
