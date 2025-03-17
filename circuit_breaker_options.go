package transport

import (
	"log/slog"

	"github.com/sony/gobreaker/v2"
)

type CircuitBreakerOption func(*circuitBreakerConfig) *circuitBreakerConfig

func CircuitBreakerOptionMatcherConfig(config MatcherConfig) CircuitBreakerOption {
	return func(c *circuitBreakerConfig) *circuitBreakerConfig {
		c.MatcherConfig = config
		return c
	}
}

func CircuitBreakerOptionLogger(logger *slog.Logger) CircuitBreakerOption {
	return func(c *circuitBreakerConfig) *circuitBreakerConfig {
		c.logger = logger
		return c
	}
}

func CircuitBreakerOptionBreakerConfig(config gobreaker.Settings) CircuitBreakerOption {
	return func(c *circuitBreakerConfig) *circuitBreakerConfig {
		c.breakerConfig = config
		return c
	}
}
