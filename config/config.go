// Package config defines globally used interfaces and structs.
package config

import (
	"context"
	"net/http"
	"time"
)

var (
	// ErrBadRequest defines an error which triggers a StatusBadRequest (http 400) to be sent
	ErrBadRequest error

	// SupportedLangs defines the languages supported by the proxied services.
	// It should be set once the program starts.
	SupportedLangs []string
)

type (
	// Env represents a collection of interfaces required for the handlers.
	Env struct {
		Auth Authenticator
		DB   Datastore
	}

	// RegistrationReqBody represents the expected request body from the /register route
	RegistrationReqBody struct {
		Email     string `json:"email"`
		Pass      string `json:"pass"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	// LoginReqBody represents the expected request body from the /login route
	LoginReqBody struct {
		Email string `json:"email"`
		Pass  string `json:"pass"`
	}
)

type (
	// Authenticator defines functions for authentication using JWTs
	Authenticator interface {
		CreateAuthCookie(uid uint64, expire time.Time) (*http.Cookie, error)
		ValidateAuthCookie(c *http.Cookie) error
		UID(c *http.Cookie) (uint64, error)
		ExpiresAt(c *http.Cookie) (time.Time, error)
	}

	// Datastore defines functions a datastore has to implement.
	Datastore interface {
		Login(ctx context.Context, body LoginReqBody) (uid uint64, pass string, lang string, err error)
		Register(ctx context.Context, body RegistrationReqBody) error
	}
)

// DefaultExpTime returns the default expiration time when an authentication token should expire.
func DefaultExpTime() time.Time {
	return time.Now().Add(time.Minute * 2)
}
