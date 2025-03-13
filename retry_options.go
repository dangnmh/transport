package transport

type RetryOption func(*retryConfig) *retryConfig

func RetryOptionOnError(enable bool) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.RetryOnError = enable
		return c
	}
}

func RetryOptionOnStatus(status []int) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.OnStatus = status
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

// * mean all example GET|/v1/user/get
func RetryOptionWhiteListPaths(paths []string) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.WhiteListPaths = paths
		return c
	}
}

// * mean all example GET|/v1/user/get
func RetryOptionBlackListPaths(paths []string) RetryOption {
	return func(c *retryConfig) *retryConfig {
		c.BlackListPaths = paths
		return c
	}
}
