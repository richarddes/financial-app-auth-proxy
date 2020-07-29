package auth

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// CreateAuthCookie returns a new JWT authenticaton token with the id set to uid and
// the expiration time set to expire.
// It returns an error when the uid < 1 or the specified expiration time already passed or the 
// expiration time is more than 5 minutes away.
func (auth *Auth) CreateAuthCookie(uid uint64, expire time.Time) (*http.Cookie, error) {
	if uid < 1 {
		return nil, errors.New("The uid cannot be smaller than 1")
	}

	if time.Now().After(expire) {
		return nil, errors.New("The expiration date has to be in the future")
	}

	if expire.Sub(time.Now()) >= time.Minute*5 {
		return nil, errors.New("The expiration time cannot be more than 5 minutes in the future")
	}

	cl := &jwt.StandardClaims{
		ExpiresAt: expire.Unix(),
		Id:        strconv.FormatUint(uid, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)

	tokenStr, err := token.SignedString(auth.jwtKey)
	if err != nil {
		return nil, err
	}

	c := &http.Cookie{
		Name:     "auth_token",
		Path:     "/api",
		Expires:  expire,
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	return c, nil
}
