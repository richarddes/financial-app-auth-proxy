package auth

import (
	"net/http"
	"strconv"
)

// UID returns the uid saved in an authentication token.
func (auth *Auth) UID(c *http.Cookie) (uint64, error) {
	cl, err := auth.authCookieClaims(c)
	if err != nil {
		return 0, err
	}

	uid, err := strconv.ParseUint(cl.Id, 10, 64)
	if err != nil {
		return 0, err
	}

	return uid, nil
}
