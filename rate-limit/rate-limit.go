package ratelimit

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ulule/limiter"
)

// Middleware is the middleware for basic http.Handler.
type Middleware struct {
	Limiter            *limiter.Limiter
	OnError            ErrorHandler
	OnLimitReached     LimitReachedHandler
	TrustForwardHeader bool
}

// NewMiddleware return a new instance of a basic HTTP middleware.
func NewMiddleware(limiter *limiter.Limiter, options ...Option) *Middleware {
	middleware := &Middleware{
		Limiter:        limiter,
		OnError:        DefaultErrorHandler,
		OnLimitReached: DefaultLimitReachedHandler,
	}

	for _, option := range options {
		option.apply(middleware)
	}

	return middleware
}

// Handler the middleware handler.
func (middleware *Middleware) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context, err := middleware.Limiter.Get(r.Context(), limiter.GetIPKey(r, middleware.TrustForwardHeader))
		if err != nil {
			middleware.OnError(w, r, err)
			return
		}

		w.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		w.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		w.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			middleware.OnLimitReached(w, r)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Option is used to define Middleware configuration.
type Option interface {
	apply(*Middleware)
}

type option func(*Middleware)

func (o option) apply(middleware *Middleware) {
	o(middleware)
}

// ErrorHandler is an handler used to inform when an error has occurred.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// WithErrorHandler will configure the Middleware to use the given ErrorHandler.
func WithErrorHandler(handler ErrorHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnError = handler
	})
}

// DefaultErrorHandler is the default ErrorHandler used by a new Middleware.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	panic(err)
}

// LimitReachedHandler is an handler used to inform when the limit has exceeded.
type LimitReachedHandler func(w http.ResponseWriter, r *http.Request)

// WithLimitReachedHandler will configure the Middleware to use the given LimitReachedHandler.
func WithLimitReachedHandler(handler LimitReachedHandler) Option {
	return option(func(middleware *Middleware) {
		middleware.OnLimitReached = handler
	})
}

// DefaultLimitReachedHandler is the default LimitReachedHandler used by a new Middleware.
func DefaultLimitReachedHandler(w http.ResponseWriter, r *http.Request) {
	NewError("Limit exceeded", http.StatusTooManyRequests).Write(w)
}

// WithForwardHeader will configure the Middleware to trust X-Real-IP and X-Forwarded-For headers.
func WithForwardHeader(trusted bool) Option {
	return option(func(middleware *Middleware) {
		middleware.TrustForwardHeader = trusted
	})
}

type errorResponse struct {
	Status int           `json:"-"`
	Error  *errorMessage `json:"error,omitempty"`
}

type errorMessage struct {
	Message string `json:"message"`
}

// Write writes a ApplicationResposne to the given response writer encoded as JSON.
func (er *errorResponse) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(er)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(er.Status)
	_, err = w.Write(b)
	return err
}

// NewError ...NewError
func NewError(msg string, status int) *errorResponse {
	return &errorResponse{
		Status: status,
		Error: &errorMessage{
			Message: msg,
		},
	}
}
