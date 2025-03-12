package transport

import "time"

type RetryOption func(*retryConfig) *retryConfig

func RetryOptionOnStatus(status []int) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.OnStatus = status
		return c
	}
}

func RetryOptionMaxTries(max uint64) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.MaxTries = max
		return c
	}
}

func RetryOptionMaxElapsedTime(max time.Duration) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.MaxElapsedTime = max
		return c
	}
}
