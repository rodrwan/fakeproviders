package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type apiError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
}

func (ar *apiError) Error() string {
	return fmt.Sprintf("%d -> %s", ar.StatusCode, ar.Message)
}

func newError(status int, msg string) *apiError {
	return &apiError{
		StatusCode: status,
		Message:    msg,
	}
}

// Write writes an aPIError to the given response writer encoded as JSON.
func (ar *apiError) Write(w http.ResponseWriter) error {
	b, err := json.Marshal(ar)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ar.StatusCode)
	_, err = w.Write(b)
	return err
}

var (
	aVeryLongTimeAgo = time.Unix(1, 0)

	errUnauthorizedAccess = newError(http.StatusUnauthorized, "unauthorized access")
	errTokenMismatch      = newError(http.StatusUnauthorized, "invalid token")
)

const (
	authSessionContextKey = "user_session"
	authTokenCookieName   = "access_token"

	tokenTypePrefix = "Bearer "
	tokenHeaderKey  = "Authorization"
	tokenMetaKey    = "auth_token"
)

// AuthMiddleware provides a middleware to authenticate an incoming request.
type AuthMiddleware struct {
	Token string
}

// NewAuthMiddleware creates a new AuthMiddleware with the given user session service.
func NewAuthMiddleware(token string) *AuthMiddleware {
	return &AuthMiddleware{Token: token}
}

// Handle authenticate the incoming request, if the authentication process fails then an
// ErrUnauthorizedAccess is returned.
func (m *AuthMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			// router handles the OPTIONS request to obtain the list of allowed methods.
			next.ServeHTTP(w, r)
			return
		}

		token, err := parseAuthToken(r)
		if err != nil {
			errUnauthorizedAccess.Write(w)
			return
		}

		if token != m.Token {
			errTokenMismatch.Write(w)
			return
		}
		ctx := r.Context()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ServeHTTP implements a negroni compatible signature.
func (m *AuthMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	m.Handle(next).ServeHTTP(w, r)
}

func parseAuthToken(r *http.Request) (string, error) {
	header := r.Header.Get(tokenHeaderKey)
	if !strings.HasPrefix(header, tokenTypePrefix) {
		return "", errors.New("auth: no token authorization header present")
	}
	return header[len(tokenTypePrefix):], nil
}
