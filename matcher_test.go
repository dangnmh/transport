package transport

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type testMatcherCase struct {
	name           string
	config         MatcherConfig
	expectedResult bool
	method         string
	url            string
	status         int
}

func TestMatcher_Cases(t *testing.T) {
	testCases := []*testMatcherCase{
		notMatchStatus,
		notMatchStatusNOrWhileList,
		matchStatusButNotWhileList,
		matchStatusWhileListAll,
		matchStatusWhileListPath,
		matchStatusWhileListPathBlackListPath,
		matchStatusBlackListPathAll,
	}

	for _, test := range testCases {
		matcher := NewMatcher(test.config)
		url, err := url.Parse(test.url)
		require.NoError(t, err, test.name)

		value := matcher.Match(&http.Request{
			Method: test.method,
			URL:    url,
		}, test.status)

		require.Equal(t, test.expectedResult, value, test.name)
	}
}

var notMatchStatus = &testMatcherCase{
	name:           "notMatchStatus",
	config:         MatcherConfig{},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: false,
}

var notMatchStatusNOrWhileList = &testMatcherCase{
	name: "notMatchStatusNOrWhileList",
	config: MatcherConfig{
		OnStatus:       []int{http.StatusOK},
		BlackListPaths: []string{ConsCharStar},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: false,
}

var matchStatusButNotWhileList = &testMatcherCase{
	name: "matchStatusButNotWhileList",
	config: MatcherConfig{
		OnStatus: []int{http.StatusServiceUnavailable},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: false,
}

var matchStatusWhileListAll = &testMatcherCase{
	name: "matchStatusWhileListAll",
	config: MatcherConfig{
		OnStatus:       []int{http.StatusServiceUnavailable},
		WhiteListPaths: []string{ConsCharStar},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: true,
	status:         http.StatusServiceUnavailable,
}

var matchStatusWhileListPath = &testMatcherCase{
	name: "matchStatusWhileListPath",
	config: MatcherConfig{
		OnStatus:       []int{http.StatusServiceUnavailable},
		WhiteListPaths: []string{CombinePath(defaultMethod, defaultPath)},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: true,
	status:         http.StatusServiceUnavailable,
}

var matchStatusWhileListPathBlackListPath = &testMatcherCase{
	name: "matchStatusWhileListPathBlackListPath",
	config: MatcherConfig{
		OnStatus:       []int{http.StatusServiceUnavailable},
		WhiteListPaths: []string{CombinePath(defaultMethod, defaultPath)},
		BlackListPaths: []string{CombinePath(defaultMethod, defaultPath)},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: false,
	status:         http.StatusServiceUnavailable,
}

var matchStatusBlackListPathAll = &testMatcherCase{
	name: "matchStatusBlackListPathAll",
	config: MatcherConfig{
		OnStatus:       []int{http.StatusServiceUnavailable},
		WhiteListPaths: []string{CombinePath(defaultMethod, defaultPath)},
		BlackListPaths: []string{ConsCharStar},
	},
	method:         defaultMethod,
	url:            defaultURL,
	expectedResult: false,
	status:         http.StatusServiceUnavailable,
}
