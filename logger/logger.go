package logger

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// LoggerDefaultDateFormat is the format used for date by the default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// Logger provides a middleware to log an incoming request.
type Logger struct {
	name   string
	logger *logrus.Logger
	before func(*logrus.Entry, *http.Request, string) *logrus.Entry
	after  func(*logrus.Entry, ResponseWriter, time.Time, string) *logrus.Entry
}

// NewLogger creates a new AuthMiddleware with the given user session service.
func NewLogger(svc string) *Logger {
	log.SetFlags(0)
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	return &Logger{
		name:   svc,
		logger: logger,
		before: DefaultBefore,
		after:  DefaultAfter,
	}
}

// Handle print incoming request
func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		entry := logrus.NewEntry(l.logger)
		entry = l.before(entry, r, l.name)

		entry.Info("starting request")

		if r.Method == http.MethodOptions {
			// router handles the OPTIONS request to obtain the list of allowed methods.
			res := w.(ResponseWriter)
			l.after(entry, res, start, l.name).Info("request completed")
			next.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)

		res := w.(ResponseWriter)

		l.after(entry, res, start, l.name).Info("request completed")
	})
}

// ServeHTTP implements a negroni compatible signature.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.Handle(next).ServeHTTP(w, r)
}

// DefaultBefore print log before request
func DefaultBefore(entry *logrus.Entry, r *http.Request, name string) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"service": name,
		"method":  r.Method,
		"URL":     r.URL,
	})
}

// DefaultAfter print log after request
func DefaultAfter(entry *logrus.Entry, res ResponseWriter, start time.Time, name string) *logrus.Entry {
	return entry.WithFields(logrus.Fields{
		"service":     name,
		"status_code": res.Status(),
		"status":      http.StatusText(res.Status()),
		"took":        fmt.Sprintf("%.2fs", time.Since(start).Seconds()),
	})
}
