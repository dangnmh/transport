package transport

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

type logTransport struct {
	tp      http.RoundTripper
	config  *logConfig
	matcher Matcher
	logger  *slog.Logger
}

type logConfig struct {
	*MatcherConfig
	level               slog.Level
	logHeaders          bool
	logLatency          bool
	logReqBody          bool
	logResBody          bool
	logQueryParams      bool     // Log query parameters separately
	maxLogBodySize      int      // Max size for logging body 0 mean unlimit
	redactSensitive     bool     // Redact sensitive headers like Authorization
	redactSensitiveKeys []string // Redact sensitive headers like Authorization
	statusLogLevels     map[int]slog.Level
	logger              *slog.Logger
}

var DefaultLogConfig = logConfig{
	MatcherConfig:       &DefaultMatcherConfig,
	level:               slog.LevelInfo,
	logHeaders:          true,
	logLatency:          true,
	logReqBody:          true,
	logResBody:          true,
	maxLogBodySize:      0,
	redactSensitive:     true,
	redactSensitiveKeys: []string{"Authorization", "Cookie", "Set-Cookie"},
	statusLogLevels: map[int]slog.Level{
		500: slog.LevelError,
		400: slog.LevelWarn,
		200: slog.LevelInfo,
	},
}

func NewTransportLog(tp http.RoundTripper, opts ...LogOption) http.RoundTripper {
	cfg := DefaultLogConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	cfg.redactSensitiveKeys = StringLowers(cfg.redactSensitiveKeys)

	return &logTransport{
		tp:      tp,
		config:  &cfg,
		matcher: NewMatcher(*cfg.MatcherConfig),
	}
}

func (lt *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	logField := lt.buildLogRequestFields(req)

	res, err := lt.tp.RoundTrip(req)
	if err != nil {
		lt.logger.ErrorContext(req.Context(), "Request failed", slog.String("error", err.Error()))
		return nil, err
	}

	if !lt.matcher.Match(req, res.StatusCode) {
		return res, nil
	}

	logField = append(logField, lt.buildLogResponseFields(res, time.Since(start))...)
	lt.logger.LogAttrs(req.Context(), lt.getLogLevel(res.StatusCode), "HTTP Request", logField...)

	return res, nil
}

func (lt *logTransport) buildLogRequestFields(req *http.Request) []slog.Attr {
	fields := lt.commonFields(req)

	if lt.config.logQueryParams {
		fields = append(fields, slog.Any("query_params", req.URL.Query()))
	}

	if lt.config.logHeaders {
		fields = append(fields, slog.Any("headers", lt.sanitizeHeaders(req.Header)))
	}

	if lt.config.logReqBody && req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		if lt.config.maxLogBodySize != 0 && len(body) > lt.config.maxLogBodySize {
			fields = append(fields, slog.String("body", string(body[:lt.config.maxLogBodySize])+"..."))
		} else {
			fields = append(fields, slog.String("body", string(body)))
		}
	}

	return fields
}

func (lt *logTransport) buildLogResponseFields(res *http.Response, latency time.Duration) []slog.Attr {
	fields := lt.commonFields(res.Request)
	fields = append(fields, slog.Int("status", res.StatusCode))

	if lt.config.logLatency {
		fields = append(fields, slog.Duration("latency", latency))
	}

	if lt.config.logResBody && res.Body != nil {
		body, _ := io.ReadAll(res.Body)
		res.Body = io.NopCloser(bytes.NewBuffer(body))
		if lt.config.maxLogBodySize != 0 && len(body) > lt.config.maxLogBodySize {
			fields = append(fields, slog.String("body", string(body[:lt.config.maxLogBodySize])+"..."))
		} else {
			fields = append(fields, slog.String("body", string(body)))
		}
	}

	return fields
}

func (lt *logTransport) commonFields(req *http.Request) []slog.Attr {
	fields := []slog.Attr{
		slog.String("method", req.Method),
		slog.String("url", req.URL.Path),
		slog.String("remote_addr", req.RemoteAddr),
	}

	if reqID := req.Header.Get("X-Request-ID"); reqID != "" {
		fields = append(fields, slog.String("request_id", reqID))
	}

	if traceID := req.Header.Get("X-Trace-ID"); traceID != "" {
		fields = append(fields, slog.String("trace_id", traceID))
	}

	return fields
}

// getLogLevel returns the appropriate log level for the status code
func (lt *logTransport) getLogLevel(status int) slog.Level {
	if level, exists := lt.config.statusLogLevels[status]; exists {
		return level
	}
	if status >= 500 {
		return slog.LevelError
	}
	if status >= 400 {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

func (lt *logTransport) sanitizeHeaders(headers http.Header) map[string][]string {
	if !lt.config.redactSensitive {
		return headers
	}

	sensitiveKeys := lt.config.redactSensitiveKeys
	sanitized := make(map[string][]string)
	for key, values := range headers {
		if slices.Contains(sensitiveKeys, strings.ToLower(key)) {
			sanitized[key] = []string{"[REDACTED]"}
		} else {
			sanitized[key] = values
		}
	}
	return sanitized
}
