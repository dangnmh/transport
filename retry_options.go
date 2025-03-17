package transport

type RetryOption func(*retryConfig) *retryConfig

func RetryOptionOnError(enable bool) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.RetryOnError = enable
		return c
	}
}

func RetryOptionMaxTries(max uint64) RetryOption {
	return func(c *retryConfig) *retryConfig {
		if max > 0 {
			max -= 1
		}

		c.MaxTries = max
		return c
	}
}

func RetryOptionMatcherConfig(config MatcherConfig) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.MatcherConfig = config
		return c
	}
}
