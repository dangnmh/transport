package transport

import (
	"bytes"
	"io"
	"net/http"
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
