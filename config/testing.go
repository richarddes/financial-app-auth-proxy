package config

import (
	"context"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type (
	mockAuth struct{}

	mockDB struct {
		store map[string]string
	}
)

func (auth *mockAuth) CreateAuthCookie(uid uint64, expire time.Time) (*http.Cookie, error) {
	c := &http.Cookie{
		Name:     "auth_token",
		Path:     "/api",
		Expires:  expire,
		Value:    "some-information",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	return c, nil
}

func (auth *mockAuth) ValidateAuthCookie(c *http.Cookie) error {
	return nil
}

func (auth *mockAuth) UID(c *http.Cookie) (uint64, error) {
	return 0, nil
}

func (auth *mockAuth) ExpiresAt(c *http.Cookie) (time.Time, error) {
	return time.Unix(0, 0), nil
}

func (db *mockDB) Login(ctx context.Context, body LoginReqBody) (uid uint64, pass string, lang string, err error) {
	pwd := db.store[body.Email]
	if pwd == "" {
		return 0, "", "", errors.New("No user with this email exists")
	}

	return 1, pwd, "en", nil
}

func (db *mockDB) Register(ctx context.Context, body RegistrationReqBody) error {
	pwd, err := bcrypt.GenerateFromPassword([]byte(body.Pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	db.store[body.Email] = string(pwd)

	return nil
}

// NewMockEnv returns a new Env with mock values instead of production values.
func NewMockEnv() *Env {
	env := new(Env)

	db := new(mockDB)
	db.store = make(map[string]string)

	auth := new(mockAuth)

	env.DB = db
	env.Auth = auth

	return env
}
