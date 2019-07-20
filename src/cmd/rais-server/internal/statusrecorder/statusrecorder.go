package statusrecorder

import "net/http"

// StatusRecorder wraps an http.ResponseWriter.  It intercepts WriteHeader
// calls so we can record the status code for logging purposes.
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

// New initializes the fake writer to a status of 200 - if a status isn't
// explicitly written, the http library will default to 200 but we won't have
// captured it if we don't also default it
func New(w http.ResponseWriter) *StatusRecorder {
	return &StatusRecorder{w, http.StatusOK}
}

// WriteHeader stores and then passes the code down to the real writer
func (rec *StatusRecorder) WriteHeader(code int) {
	rec.Status = code
	rec.ResponseWriter.WriteHeader(code)
}
