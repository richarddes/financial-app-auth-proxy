// Package auth provides function for using JWT authentication tokens.
package auth

import (
	"errors"
	"net/http"

	"auth-proxy/config"

	"github.com/dgrijalva/jwt-go"
)

// Auth defines a new Authenticator
type Auth struct {
	config.Authenticator
	jwtKey []byte
}

// New returns a new instance of the Auth authenticator with the jwt set.
func New(jwtKey string) (*Auth, error) {
	if jwtKey == "" {
		return nil, errors.New("The jwt key can't be an empty string")
	}

	auth := new(Auth)
	auth.jwtKey = []byte(jwtKey)

	return auth, nil
}

// authCookieClaims is a helper function for the ExpiresAt and UID function
func (auth *Auth) authCookieClaims(c *http.Cookie) (*jwt.StandardClaims, error) {
	tokenStr := c.Value
	cl := &jwt.StandardClaims{}

	token, err := jwt.ParseWithClaims(tokenStr, cl, func(token *jwt.Token) (interface{}, error) {
		return auth.jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return cl, nil
}
