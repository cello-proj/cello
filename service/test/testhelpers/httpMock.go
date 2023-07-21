package testhelpers

import (
	"bytes"
	"net/http"
)

type mockResponseWriter struct {
	StatusCode int           // the HTTP response code from WriteHeader
	HeaderMap  http.Header   // the HTTP response headers
	Body       *bytes.Buffer // if non-nil, the bytes.Buffer to append written data to
}

func (w *mockResponseWriter) Header() http.Header {
	return w.HeaderMap
}

func (w *mockResponseWriter) Write(buf []byte) (int, error) {
	if w.Body != nil {
		return w.Body.Write(buf)
	}
	if w.StatusCode == 0 {
		w.StatusCode = http.StatusOK
	}
	return len(buf), nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
}

func NewMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
	}
}
