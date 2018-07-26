package logger

import (
	"net/http"
)

// responseWriter wraps a standard http.ResponseWriter
// so we can store the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	return &responseWriter{rw, http.StatusOK}
}

// WriteHeader ...
func (rw *responseWriter) WriteHeader(statusCode int) {
	// Store the status code
	rw.status = statusCode
	// Write the status code onward.
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Status() int {
	return rw.status
}
