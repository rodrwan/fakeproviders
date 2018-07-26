package logger

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// responseWriter wraps a standard http.ResponseWriter
// so we can store the status code.
type responseWriter struct {
	http.ResponseWriter
	status      int
	beforeFuncs []beforeFunc
}

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	// Status returns the status code of the response or 0 if the response has
	// not been written
	Status() int
	// Written returns whether or not the ResponseWriter has been written.
	Written() bool
	// Before allows for a function to be called before the ResponseWriter has been written to. This is
	// useful for setting headers or any other operations that must happen before a response has been written.
	Before(func(ResponseWriter))
}

type beforeFunc func(ResponseWriter)

func newMyResponseWriter(res http.ResponseWriter) ResponseWriter {
	mrw := &responseWriter{
		ResponseWriter: res,
	}

	if _, ok := res.(http.CloseNotifier); ok {
		return &responseWriterCloseNotifer{mrw}
	}

	return mrw
}

// Status Give a way to get the status
func (rw *responseWriter) Status() int {
	return rw.status
}

// Header Satisfy the http.ResponseWriter interface
func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Written() bool {
	return rw.status != 0
}

// Before ...
func (rw *responseWriter) Before(before func(ResponseWriter)) {
	rw.beforeFuncs = append(rw.beforeFuncs, before)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// WriteHeader ...
func (rw *responseWriter) WriteHeader(statusCode int) {
	// Store the status code
	rw.status = statusCode
	rw.callBefore()
	// Write the status code onward.
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) callBefore() {
	for i := len(rw.beforeFuncs) - 1; i >= 0; i-- {
		rw.beforeFuncs[i](rw)
	}
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if ok {
		if !rw.Written() {
			// The status will be StatusOK if WriteHeader has not been called yet
			rw.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

type responseWriterCloseNotifer struct {
	*responseWriter
}

func (rw *responseWriterCloseNotifer) CloseNotify() <-chan bool {
	return rw.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
