package transport

import "log/slog"

type LogOption func(*logConfig) *logConfig

func LogOptionLevel(level slog.Level) LogOption {
	return func(c *logConfig) *logConfig {
		c.level = level
		return c
	}
}

func LogOptionHeaders(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.logHeaders = enable
		return c
	}
}

func LogOptionLatency(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.logLatency = enable
		return c
	}
}

func LogOptionReqBody(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.logReqBody = enable
		return c
	}
}

func LogOptionResBody(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.logResBody = enable
		return c
	}
}

func LogOptionQueryParams(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.logQueryParams = enable
		return c
	}
}

func LogOptionMaxLogBodySize(max int) LogOption {
	return func(c *logConfig) *logConfig {
		c.maxLogBodySize = max
		return c
	}
}

func LogOptionLogger(logger *slog.Logger) LogOption {
	return func(c *logConfig) *logConfig {
		c.logger = logger
		return c
	}
}

func LogOptionRedactSensitive(enable bool) LogOption {
	return func(c *logConfig) *logConfig {
		c.redactSensitive = enable
		return c
	}
}

func LogOptionRedactSensitiveKeys(keys []string) LogOption {
	return func(c *logConfig) *logConfig {
		c.redactSensitiveKeys = keys
		return c
	}
}

func LogOptionStatusLogLevels(statusLogLevels map[int]slog.Level) LogOption {
	return func(c *logConfig) *logConfig {
		c.statusLogLevels = statusLogLevels
		return c
	}
}
