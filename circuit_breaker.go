package transport

import (
	"errors"
	"net/http"

	"log/slog"

	"github.com/sony/gobreaker/v2"
)

type circuitBreakerTransport struct {
	tp      http.RoundTripper
	breaker *gobreaker.CircuitBreaker[*http.Response]
	logger  *slog.Logger
	matcher Matcher
}

type circuitBreakerConfig struct {
	MatcherConfig
	logger        *slog.Logger
	breakerConfig gobreaker.Settings
}

var DefaultCircuitBreakerConfig = circuitBreakerConfig{
	MatcherConfig: DefaultMatcherConfig,
	logger:        defaultLogger,
}

// NewCircuitBreakerTransport wraps a RoundTripper with a circuit breaker.
func NewCircuitBreakerTransport(tp http.RoundTripper, opts ...CircuitBreakerOption) http.RoundTripper {
	cfg := DefaultCircuitBreakerConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return &circuitBreakerTransport{
		tp:      tp,
		logger:  cfg.logger,
		breaker: gobreaker.NewCircuitBreaker[*http.Response](cfg.breakerConfig),
		matcher: NewMatcher(cfg.MatcherConfig),
	}
}

func (cbt *circuitBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !cbt.matcher.MatchPath(req) {
		return cbt.tp.RoundTrip(req)
	}

	result, err := cbt.breaker.Execute(func() (*http.Response, error) {
		res, err := cbt.tp.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if cbt.matcher.Match(req, res.StatusCode) {
			return nil, errors.New("server error")
		}

		return res, nil
	})

	if err != nil {
		cbt.logger.WarnContext(req.Context(), "Circuit breaker triggered", slog.String("error", err.Error()))
		return nil, err
	}

	return result, nil
}
