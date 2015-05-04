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

func NewResponseWriter() *ResponseWriter {
	return &ResponseWriter{
		StatusCode: -1,
		Headers:    make(http.Header),
		Output:     make([]byte, 0),
	}
}

func (rw *ResponseWriter) Header() http.Header {
	return rw.Headers
}

func (rw *ResponseWriter) Write(b []byte) (int, error) {
	lenCurr, lenNew := len(rw.Output), len(b)
	newOut := make([]byte, lenCurr+lenNew)
	copy(newOut, rw.Output)
	copy(newOut[lenCurr:], b)
	rw.Output = newOut
	return len(b), nil
}

func (rw *ResponseWriter) WriteHeader(s int) {
	rw.StatusCode = s
}
