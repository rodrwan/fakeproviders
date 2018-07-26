package logger

import "net/http"

// MyResponseWriter wraps a standard http.ResponseWriter
// so we can store the status code.
type MyResponseWriter struct {
	status int
	http.ResponseWriter
}

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	// Status returns the status code of the response or 0 if the response has
	// not been written
	Status() int
}

func newMyResponseWriter(res http.ResponseWriter) ResponseWriter {
	mrw := &MyResponseWriter{
		ResponseWriter: res,
	}

	if _, ok := res.(http.CloseNotifier); ok {
		return &responseWriterCloseNotifer{mrw}
	}

	return mrw
}

// Status Give a way to get the status
func (w MyResponseWriter) Status() int {
	return w.status
}

// Header Satisfy the http.ResponseWriter interface
func (w *MyResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *MyResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// WriteHeader ...
func (w *MyResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.status = statusCode
	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}

type responseWriterCloseNotifer struct {
	*MyResponseWriter
}

func (rw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
