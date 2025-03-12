package transport

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	name             string
	errs             []error
	attempts         int
	reqBody          []byte
	responses        []*http.Response
	options          []RetryOption
	expectedError    bool
	expectedResponse bool
	expectedAttempts int
	expectedStatus   int
	method           string
	url              string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Println("counter", m.attempts)
	if m.attempts < len(m.errs) && m.errs[m.attempts] != nil {
		m.attempts++
		return nil, m.errs[m.attempts-1]
	}
	if m.attempts < len(m.responses) {
		resp := m.responses[m.attempts]
		fmt.Println(m.attempts, resp.StatusCode, "11111111111")
		m.attempts++
		return resp, nil
	}
	return nil, errors.New("unexpected request")
}

func TestRetryTransport_Cases(t *testing.T) {
	testCases := []*mockRoundTripper{
		// noRetryOnError,
		// successRetryOnError,
		// noRetryOnSuccess,
		// retryOnValidStatus,
		// noRetryOnStatus,
		// noRetryBlackListAll,
		// noBlackList,
		// notWhiteListRetry,
		// successWhiteListRetry,
		// noRetryCauseMissWhiteListRetry,
		// noRetryCausWhiteListBlackListRetry,
		successOnMaxRetry,
		// failedOnMaxRetry,
	}

	for _, testCase := range testCases {
		client := &http.Client{
			Transport: NewTransportRetry(testCase, testCase.options...),
		}

		testName := testCase.name

		req, err := http.NewRequest(testCase.method, testCase.url, bytes.NewReader(testCase.reqBody))
		require.NoError(t, err, testCase.name)
		resp, err := client.Do(req)

		require.Equal(t, testCase.expectedError, err != nil, testName)
		require.Equal(t, testCase.expectedAttempts, testCase.attempts, testName)
		require.Equal(t, testCase.expectedResponse, resp != nil, testName)
		if testCase.expectedResponse {
			require.Equal(t, testCase.expectedStatus, resp.StatusCode, testName)
		}
	}
}

var temporaryNetworkError = errors.New("temporary network error")
var defaultMethod = http.MethodGet
var defaultDomain = "http://example.com"
var defaultPath = "/v1/api"
var defaultURL = defaultDomain + defaultPath

var noRetryOnError = &mockRoundTripper{
	name: "noRetryOnError",
	errs: []error{
		temporaryNetworkError,
	},
	responses: []*http.Response{
		nil,
	},
	expectedAttempts: 1,
	expectedError:    true,
	expectedResponse: false,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionOnError(false)},
}

var successRetryOnError = &mockRoundTripper{
	name: "successRetryOnError",
	errs: []error{
		temporaryNetworkError,
		nil,
	},
	responses: []*http.Response{
		nil,
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 2,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionOnError(true)},
	expectedStatus:   http.StatusOK,
}

var noRetryOnSuccess = &mockRoundTripper{
	name: "noRetryOnSuccess",
	errs: []error{
		nil,
	},
	responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{},
	expectedStatus:   http.StatusOK,
}

var retryOnValidStatus = &mockRoundTripper{
	name: "retryOnValidStatus",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusOK, // Second attempt succeeds
			Body:       io.NopCloser(strings.NewReader("OK")),
		},
	},
	expectedAttempts: 2,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{},
	expectedStatus:   http.StatusOK,
}

var noRetryOnStatus = &mockRoundTripper{
	name: "noRetryOnStatus",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionOnStatus([]int{})},
	expectedStatus:   http.StatusServiceUnavailable,
}

var noBlackList = &mockRoundTripper{
	name: "noBlackList",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
		},
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 2,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionBlackListPaths([]string{})},
	expectedStatus:   http.StatusOK,
}

var noRetryBlackListAll = &mockRoundTripper{
	name: "noRetryBlackListAll",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionBlackListPaths([]string{ConsCharStar})},
	expectedStatus:   http.StatusServiceUnavailable,
}

var notWhiteListRetry = &mockRoundTripper{
	name: "notWhiteListRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionWhiteListPaths([]string{})},
	expectedStatus:   http.StatusServiceUnavailable,
}

var successWhiteListRetry = &mockRoundTripper{
	name: "successWhiteListRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 2,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionWhiteListPaths([]string{fmt.Sprintf("%s%s%s", defaultMethod, ConsCharVerticalBar, defaultPath)})},
	expectedStatus:   http.StatusOK,
}

var noRetryCauseMissWhiteListRetry = &mockRoundTripper{
	name: "noRetryCauseMissWhiteListRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options:          []RetryOption{RetryOptionWhiteListPaths([]string{fmt.Sprintf("%s%s%s", defaultMethod, ConsCharVerticalBar, defaultPath+"/invalid")})},
	expectedStatus:   http.StatusServiceUnavailable,
}

var noRetryCausWhiteListBlackListRetry = &mockRoundTripper{
	name: "noRetryCausWhiteListBlackListRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 1,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options: []RetryOption{
		RetryOptionWhiteListPaths([]string{fmt.Sprintf("%s%s%s", defaultMethod, ConsCharVerticalBar, defaultPath)}),
		RetryOptionBlackListPaths([]string{fmt.Sprintf("%s%s%s", defaultMethod, ConsCharVerticalBar, defaultPath)}),
	},
	expectedStatus: http.StatusServiceUnavailable,
}

var successOnMaxRetry = &mockRoundTripper{
	name: "successOnMaxRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusOK,
		},
	},
	expectedAttempts: 2,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options: []RetryOption{
		RetryOptionMaxTries(4),
	},
	expectedStatus: http.StatusOK,
}

var failedOnMaxRetry = &mockRoundTripper{
	name: "failedOnMaxRetry",
	errs: []error{},
	responses: []*http.Response{
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
		{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(strings.NewReader("Service Unavailable")),
		},
	},
	expectedAttempts: 4,
	expectedError:    false,
	expectedResponse: true,
	method:           defaultMethod,
	url:              defaultURL,
	reqBody:          []byte{},
	options: []RetryOption{
		RetryOptionMaxTries(3),
	},
	expectedStatus: http.StatusServiceUnavailable,
}
