package logger

import (
	"log"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// LoggerDefaultDateFormat is the format used for date by the default Logger instance.
var LoggerDefaultDateFormat = time.RFC3339

// Logger provides a middleware to log an incoming request.
type Logger struct {
	Service string
	logger  *logrus.Logger
}

// NewLogger creates a new AuthMiddleware with the given user session service.
func NewLogger(svc string) *Logger {
	log.SetFlags(0)
	logger := logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	return &Logger{
		Service: svc,
		logger:  logger,
	}
}

// Handle print incoming request
func (l *Logger) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		l.logger.Infof("%s: %s | started %s", r.Method, r.URL, start.Format(LoggerDefaultDateFormat))
		if r.Method == http.MethodOptions {
			// router handles the OPTIONS request to obtain the list of allowed methods.
			l.logger.Infof("%s: %s | took %.2fs", r.Method, r.URL, time.Since(start).Seconds())
			next.ServeHTTP(w, r)
			return
		}

		l.logger.Infof("%s: %s | took %.2fs", r.Method, r.URL, time.Since(start).Seconds())
		next.ServeHTTP(w, r)
	})
}

// ServeHTTP implements a negroni compatible signature.
func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	l.Handle(next).ServeHTTP(w, r)
}
