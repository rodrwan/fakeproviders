// Package jwt implements sigiriya.SessionService using JWT tokens.
//
//  - Validation Token keys:
//   * standard: jti, iat, sub, exp, iss
//  - Authentication Token kys:
//   * standard: jti, iat, sub, exp, iss
//   * custom: id, email, host, created_at, updated_at
package jwt

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	tokenDuration   = 72
	expireOffset    = 3600
	tokenIDnumBytes = 32
)

type sessionClaims struct {
	jwt.StandardClaims

	// Custom claims used to store user session.
	ID        string `json:"id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Origin    string `json:"origin,omitempty"`
	Email     string `json:"email,omitempty"`
	CreatedAt int64  `json:"created_at,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

// Session represents an user session.
type Session struct {
	ID string `json:"id" db:"id"`

	UserID string `json:"user_id" db:"user_id"`
	Email  string `json:"email" db:"email"`
	Origin string `json:"origin" db:"origin"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewSession creates a new user session.
func NewSession(username, uuid, origin string) (*Session, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	iat := time.Now()
	id := base64.StdEncoding.EncodeToString(b)

	return &Session{
		ID:        id,
		UserID:    uuid,
		Email:     username,
		Origin:    origin,
		CreatedAt: iat,
		UpdatedAt: iat,
	}, nil
}

// SessionCredentials represents credentials of an user session.
type SessionCredentials struct {
	AuthToken string `json:"auth_token,omitempty"`
}

func (sc *sessionClaims) Session() *Session {
	return &Session{
		ID:        sc.ID,
		Email:     sc.Email,
		UserID:    sc.UserID,
		Origin:    sc.Origin,
		CreatedAt: time.Unix(sc.CreatedAt, 0),
		UpdatedAt: time.Unix(sc.UpdatedAt, 0),
	}
}

// SessionService implements dale.SessionService using JWT tokens.
type SessionService struct {
	SecretKey []byte
	MaxAge    time.Duration
}

// Session validates and returns the user session associated with the given
// credentials.
func (uss *SessionService) Session(ctx context.Context, sc *SessionCredentials) (*Session, error) {
	authClaims, err := uss.parseTokens(sc.AuthToken)
	if err != nil {
		return nil, err
	}

	// Get data from auth token.
	sess := authClaims.Session()
	return sess, nil
}

// RefreshSession validates and returns the user session associated with the
// given credentials. This method skips the validation of the expiry of the
// tokens.
// Also the associated user session is returned updated.
func (uss *SessionService) RefreshSession(ctx context.Context, c *SessionCredentials) (*Session, error) {
	authClaims, err := uss.parseTokens(c.AuthToken)
	if err != nil {
		if !isTokenExpired(err) {
			return nil, err
		}
	}

	s := authClaims.Session()
	s.UpdatedAt = time.Now()
	return s, nil
}

// HasExpired ...
func (uss *SessionService) HasExpired(ctx context.Context, c *SessionCredentials) (bool, error) {
	_, err := uss.tokenClaims(c.AuthToken)
	if err != nil {
		if isTokenExpired(err) {
			return true, err
		}

		return false, err
	}

	return false, nil
}

// CreateSession creates new credentials for the given session.
func (uss *SessionService) CreateSession(ctx context.Context, us *Session) (*SessionCredentials, error) {
	return uss.sessionCredentials(us)
}

// UpdateSession creates new credentials for the given session.
func (uss *SessionService) UpdateSession(ctx context.Context, us *Session) (*SessionCredentials, error) {
	return uss.sessionCredentials(us)
}

func (uss *SessionService) sessionCredentials(us *Session) (*SessionCredentials, error) {
	id, err := generateRandomToken(tokenIDnumBytes)
	if err != nil {
		return nil, err
	}

	iat := time.Now()
	exp := iat.Add(uss.MaxAge)
	stdClms := jwt.StandardClaims{
		Id:        id,
		Issuer:    us.Origin,
		Subject:   us.Email,
		IssuedAt:  iat.Unix(),
		ExpiresAt: exp.Unix(),
	}

	authToken, err := uss.tokenString(&sessionClaims{
		StandardClaims: stdClms,
		ID:             id,
		UserID:         us.UserID,
		Email:          us.Email,
		Origin:         us.Origin,
		CreatedAt:      us.CreatedAt.Unix(),
		UpdatedAt:      us.UpdatedAt.Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &SessionCredentials{
		AuthToken: authToken,
	}, nil
}

func (uss *SessionService) validateClaims(lhs, rhs *sessionClaims) error {
	if lhs.Id != rhs.Id {
		return errors.New("jwt: validation and authentication token jti mismatched")
	}

	if lhs.IssuedAt != rhs.IssuedAt {
		return errors.New("jwt: validation and authentication token iat mismatched")
	}

	if lhs.ExpiresAt != rhs.ExpiresAt {
		return errors.New("jwt: validation and authentication token exp mismatched")
	}

	if lhs.Subject != rhs.Subject {
		return errors.New("jwt: validation and authentication token sub mismatched")
	}

	if lhs.Issuer != rhs.Issuer {
		return errors.New("jwt: validation and authentication token iss mismatched")
	}

	return nil
}

func (uss *SessionService) parseTokens(authToken string) (*sessionClaims, error) {
	authClaims, authErr := uss.tokenClaims(authToken)

	var err error
	if authErr != nil {
		err = authErr
	}

	return authClaims, err
}

func (uss *SessionService) tokenClaims(tokenStr string) (*sessionClaims, error) {
	claims := &sessionClaims{}
	tkn := strings.TrimSpace(tokenStr)
	token, err := jwt.ParseWithClaims(tkn, claims, uss.verifySigningMethod)
	if err != nil {
		return nil, err
	}

	if c, ok := token.Claims.(*sessionClaims); ok {
		claims = c
	}

	return claims, nil
}

func isTokenExpired(err error) bool {
	e, ok := err.(*jwt.ValidationError)
	if !ok {
		return false
	}
	return (e.Errors & ^jwt.ValidationErrorExpired) == 0
}

func (uss *SessionService) tokenString(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	str, err := token.SignedString(uss.SecretKey)
	return str, err
}

func (uss *SessionService) verifySigningMethod(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		return nil, err
	}

	return uss.SecretKey, nil
}

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
