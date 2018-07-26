package logger

import "net/http"

// MyResponseWriter wraps a standard http.ResponseWriter
// so we can store the status code.
type MyResponseWriter struct {
	http.ResponseWriter
	status      int
	beforeFuncs []beforeFunc
}

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	// Status returns the status code of the response or 0 if the response has
	// not been written
	Status() int
	// Before allows for a function to be called before the ResponseWriter has been written to. This is
	// useful for setting headers or any other operations that must happen before a response has been written.
	Before(func(ResponseWriter))
}

type beforeFunc func(ResponseWriter)

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
func (w *MyResponseWriter) Status() int {
	return w.status
}

// Header Satisfy the http.ResponseWriter interface
func (w *MyResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

// Before ...
func (w *MyResponseWriter) Before(before func(ResponseWriter)) {
	w.beforeFuncs = append(w.beforeFuncs, before)
}

func (w *MyResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// WriteHeader ...
func (w *MyResponseWriter) WriteHeader(statusCode int) {
	// Store the status code
	w.status = statusCode
	w.callBefore()
	// Write the status code onward.
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *MyResponseWriter) callBefore() {
	for i := len(w.beforeFuncs) - 1; i >= 0; i-- {
		w.beforeFuncs[i](w)
	}
}

type responseWriterCloseNotifer struct {
	*MyResponseWriter
}

func (rw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
