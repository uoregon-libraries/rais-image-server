// Package fakehttp provides a fake response writer for use in tests.  Further
// http package mocks will probably not be created, as I'm sure there's a more
// complete http mock library out there.
package fakehttp

import (
	"net/http"
)

// The ResponseWriter adheres to the http.ResposeWriter interface in a minimal
// way to support easier testing
type ResponseWriter struct {
	StatusCode int
	Headers    http.Header
	Output     []byte
}

// NewResponseWriter returns a ResponseWriter with a default status code of -1
func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		StatusCode: -1,
		Headers:    make(http.Header),
		Output:     make([]byte, 0),
	}
}

// Header returns the http.Header data for http.ResponseWriter compatibility
func (rw *ResponseWriter) Header() http.Header {
	return rw.Headers
}

// Write stores the given output for later testing
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	lenCurr, lenNew := len(rw.Output), len(b)
	newOut := make([]byte, lenCurr+lenNew)
	copy(newOut, rw.Output)
	copy(newOut[lenCurr:], b)
	rw.Output = newOut
	return len(b), nil
}

// WriteHeader sets up StatusCode for later testing
func (rw *ResponseWriter) WriteHeader(s int) {
	rw.StatusCode = s
}
