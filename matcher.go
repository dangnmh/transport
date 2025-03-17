package transport

import (
	"net/http"
	"slices"
)

type MatcherConfig struct {
	OnStatus       []int
	WhiteListPaths []string
	BlackListPaths []string
}

var DefaultMatcherConfig = MatcherConfig{
	OnStatus:       []int{429, 502, 503, 504},
	WhiteListPaths: []string{ConsCharStar},
	BlackListPaths: []string{},
}

type Matcher interface {
	Match(req *http.Request, statusCode int) bool
	MatchPath(req *http.Request) bool
}

func NewMatcher(cfg MatcherConfig) Matcher {
	return &cfg
}

func (m *MatcherConfig) Match(req *http.Request, statusCode int) bool {
	if !slices.Contains(m.OnStatus, statusCode) {
		return false
	}

	return m.MatchPath(req)
}

func (m *MatcherConfig) MatchPath(req *http.Request) bool {
	if slices.Contains(m.BlackListPaths, ConsCharStar) {
		return false
	}

	combinedPath := CombinePath(req.Method, req.URL.Path)
	for _, path := range m.BlackListPaths {
		if MatchesPath(path, combinedPath) {
			return false
		}
	}

	if slices.Contains(m.WhiteListPaths, ConsCharStar) {
		return true
	}

	for _, path := range m.WhiteListPaths {
		if MatchesPath(path, combinedPath) {
			return true
		}
	}

	return false
}
