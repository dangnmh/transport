package transport

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type TransportOption func(*http.Transport) *http.Transport

func TransportOptionProxy(proxy func(*http.Request) (*url.URL, error)) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.Proxy = proxy
		return t
	}
}

func TransportOptionDialContext(dialCtx func(ctx context.Context, network, addr string) (net.Conn, error)) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.DialContext = dialCtx
		return t
	}
}

func TransportOptionTLSClientConfig(config *tls.Config) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.TLSClientConfig = config
		return t
	}
}

func TransportOptionTLSHandshakeTimeout(timeout time.Duration) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.TLSHandshakeTimeout = timeout
		return t
	}
}

func TransportOptionDisableKeepAlives(disabled bool) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.DisableKeepAlives = disabled
		return t
	}
}

func TransportOptionDisableCompression(disabled bool) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.DisableCompression = disabled
		return t
	}
}

func TransportOptionMaxIdleConnsPerHost(max int) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.MaxIdleConnsPerHost = max
		return t
	}
}

func TransportOptionIdleConnTimeout(timeout time.Duration) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.IdleConnTimeout = timeout
		return t
	}
}

func TransportOptionMaxIdleConns(num int) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.MaxIdleConns = num
		return t
	}
}

func TransportOptionResponseHeaderTimeout(to time.Duration) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.ResponseHeaderTimeout = to
		return t
	}
}

func TransportOptionExpectContinueTimeout(to time.Duration) TransportOption {
	return func(t *http.Transport) *http.Transport {
		t.ExpectContinueTimeout = to
		return t
	}
}
