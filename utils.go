package transport

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func cloneRequest(req *http.Request) (*http.Request, error) {
	newReq := req.Clone(req.Context()) // Clone method copies headers and context

	// If the request has a body, we need to copy it
	if req.Body != nil && req.Body != http.NoBody {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, req.Body) // Copy body content
		if err != nil {
			return nil, err
		}

		newReq.Body = io.NopCloser(bytes.NewReader(buf.Bytes())) // Reset body
		req.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))    // Reset original body
	}

	return newReq, nil
}

func CombinePath(method, path string) string {
	return fmt.Sprintf("%s%s%s", method, ConsCharVerticalBar, path)
}

func MatchesPath(pattern, path string) bool {
	if pattern == ConsCharStar {
		return false
	}

	if strings.HasSuffix(pattern, "/*") {
		return strings.HasPrefix(path, strings.TrimSuffix(pattern, "/*"))
	}

	return pattern == path
}
