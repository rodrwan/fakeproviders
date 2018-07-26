package logger

import "net/http"

// MyResponseWriter wraps a standard http.ResponseWriter
// so we can store the status code.
type MyResponseWriter struct {
	status int
	http.ResponseWriter
}

func newMyResponseWriter(res http.ResponseWriter) *MyResponseWriter {
	return &MyResponseWriter{
		ResponseWriter: res,
	}
}

// Status Give a way to get the status
func (w MyResponseWriter) Status() int {
	return w.status
}

// Header Satisfy the http.ResponseWriter interface
func (w MyResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w MyResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// WriteHeader ...
func (w MyResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.status = statusCode

	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}
