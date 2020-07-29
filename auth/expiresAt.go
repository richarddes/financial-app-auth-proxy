package auth

import (
	"net/http"
	"time"
)

// ExpiresAt returns the expiration time saved in an authentication token.
func (auth *Auth) ExpiresAt(c *http.Cookie) (time.Time, error) {
	cl, err := auth.authCookieClaims(c)
	if err != nil {
		return time.Unix(0, 0), err
	}

	return time.Unix(cl.ExpiresAt, 0), nil
}
