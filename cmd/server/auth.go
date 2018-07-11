package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Finciero/sigiriya"
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
	errExpiredToken       = newError(http.StatusForbidden, "invalid or expired token")
	errPasswordMismatch   = newError(http.StatusForbidden, "password and password confirmation mismatch")
	errInvalidEmail       = newError(http.StatusNotFound, "user does not exists")
)

const (
	authSessionContextKey = "user_session"
	authTokenCookieName   = "access_token"

	tokenTypePrefix = "Bearer "
	tokenHeaderKey  = "Authorization"
	tokenMetaKey    = "auth_token"
)

// SessionFromContext returns the associated session with the given context.
func SessionFromContext(c context.Context) *sigiriya.Session {
	if s, ok := c.Value(authSessionContextKey).(*sigiriya.Session); ok {
		return s
	}
	return nil
}

// SessionToContext associate session with the given context.
func SessionToContext(c context.Context, us *sigiriya.Session) context.Context {
	return context.WithValue(c, authSessionContextKey, us)
}

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

		_, err := parseAuthCredentials(r)
		if err != nil {
			errUnauthorizedAccess.Write(w)
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

func parseValidationToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(authTokenCookieName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func parseAuthCredentials(r *http.Request) (*sigiriya.SessionCredentials, error) {
	authToken, err := parseAuthToken(r)
	if err != nil {
		return nil, err
	}
	validationToken, err := parseValidationToken(r)
	if err != nil {
		return nil, err
	}
	return &sigiriya.SessionCredentials{
		AuthToken:       authToken,
		ValidationToken: validationToken,
	}, nil
}
