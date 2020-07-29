package auth

import (
	"errors"
	"net/http"
	"time"
)

// ValidateAuthCookie validates an authentication token based on it's expiration time and uid.
func (auth *Auth) ValidateAuthCookie(c *http.Cookie) error {
	expiresAt, err := auth.ExpiresAt(c)
	if err != nil {
		return err
	}

	if time.Now().After(expiresAt) {
		return errors.New("The cookie's expired")
	}

	uid, err := auth.UID(c)
	if err != nil {
		return err
	}

	if uid < 1 {
		return errors.New("The uid can't be smaller than 1")
	}

	return nil
}
