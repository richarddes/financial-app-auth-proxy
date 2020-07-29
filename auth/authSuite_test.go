package auth_test

import (
	"auth-proxy/auth"
	"auth-proxy/config"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	mockJwtKey = "mock-key"
)

func createWrongAuthCookie(uid uint64, expire time.Time) (*http.Cookie, error) {
	cl := &jwt.StandardClaims{
		ExpiresAt: expire.Unix(),
		Id:        strconv.FormatUint(uid, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, cl)

	tokenStr, err := token.SignedString([]byte(mockJwtKey))
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

func AuthenticatorSuite(t *testing.T, impl config.Authenticator) {
	t.Run("test authentication cookie creation", func(t *testing.T) {
		inTwoMin := time.Now().Add(time.Minute * 2)

		cases := []struct {
			uid    uint64
			expire time.Time
			valid  bool
		}{
			{0, inTwoMin, false},
			{1, time.Now(), false},
			{2, time.Unix(0, 0), false},
			{3, time.Now().Add(time.Minute * 10), false},
			{1, inTwoMin, true},
			{150000, time.Now().Add(time.Second * 30), true},
		}

		for _, i := range cases {
			c, err := impl.CreateAuthCookie(i.uid, i.expire)
			switch {
			case err != nil && !i.valid:
				if c != nil {
					t.Fatalf("A cookie was returned even though an error occured when the uid=%d and expire=%v", i.uid, i.expire)
				}
				break

			case err != nil && i.valid:
				t.Fatalf("CreateAuthCookie returned an error: \"%v\" even though it shouldn't when the uid=%d and expire=%v", err, i.uid, i.expire)
				break

			case err == nil && !i.valid:
				t.Fatalf("CreateAuthCookie should fail when the uid=%d and expire=%v", i.uid, i.expire)
				break

			case err == nil && i.valid:
				if c == nil {
					t.Fatalf("No cookie was returned when the uid=%d and expire=%v", i.uid, i.expire)
				}

				if c.Value == "" {
					t.Fatal("The cookie has an empty value field")
				}

				if !c.Expires.Equal(i.expire) {
					t.Errorf("The expiration date in the cookie does not match the expire parameter when it's equal to %v", i.expire)
				}

				if !c.HttpOnly {
					t.Error("The cookie isn't a HttpOnly cookie")
				}

				if !c.Secure {
					t.Error("The cookie isn't a Secure cookie")
				}

				if c.SameSite != http.SameSiteStrictMode {
					t.Error("The cookies SameSite policy does not equal SameSiteStrictMode")
				}

				break
			}
		}
	})

	t.Run("test authentication cookie validation", func(t *testing.T) {
		inTwoMin := time.Now().Add(time.Minute * 2)

		crctCVals := []struct {
			uid    uint64
			expire time.Time
		}{
			{1, inTwoMin},
			{5000, inTwoMin},
			{430, inTwoMin.Add(time.Minute)},
		}

		wngCVals := []struct {
			uid    uint64
			expire time.Time
		}{
			{0, inTwoMin},
			{1, time.Unix(0, 0)},
		}

		cs := make([]*http.Cookie, len(crctCVals)+len(wngCVals))

		for i, val := range crctCVals {
			c, err := impl.CreateAuthCookie(val.uid, val.expire)
			if err != nil {
				t.Fatal(err)
			}

			cs[i] = c
		}

		for i, val := range wngCVals {
			c, err := createWrongAuthCookie(val.uid, val.expire)
			if err != nil {
				t.Fatal(err)
			}

			cs[i+len(crctCVals)] = c
		}

		for i, c := range cs {
			if i < len(crctCVals) {
				if err := impl.ValidateAuthCookie(c); err != nil {
					t.Errorf("Unexpected error: %v when uid=%d and expire%v", err, crctCVals[i].uid, crctCVals[i].expire)
				}
			} else {
				if err := impl.ValidateAuthCookie(c); err == nil {
					t.Errorf("Expected an error but got none when uid=%d and expire=%v", wngCVals[i-len(crctCVals)].uid, wngCVals[i-len(crctCVals)].expire)
				}
			}
		}
	})

	t.Run("test cookie field value reading", func(t *testing.T) {
		inTwoMin := time.Now().Add(time.Minute * 2)

		cVals := []struct {
			uid    uint64
			expire time.Time
		}{
			{1, inTwoMin},
			{50, inTwoMin.Add(-time.Second * 30)},
			{100, inTwoMin},
			{50300, inTwoMin.Add(time.Minute * 2)},
		}

		cs := make([]*http.Cookie, len(cVals))

		for i, val := range cVals {
			c, err := impl.CreateAuthCookie(val.uid, val.expire)
			if err != nil {
				t.Fatal(err)
			}

			cs[i] = c
		}

		t.Run("test uid collection", func(t *testing.T) {
			for i, c := range cs {
				uid, err := impl.UID(c)
				if err != nil {
					t.Errorf("Unexpected error: %v when uid=%d and expire=%v", err, cVals[i].uid, cVals[i].expire)
				}

				if uid != cVals[i].uid {
					t.Errorf("The uid in the cookie didn't match the specified uid")
				}
			}
		})

		t.Run("test expiration time collection", func(t *testing.T) {
			for i, c := range cs {
				expiresAt, err := impl.ExpiresAt(c)
				if err != nil {
					t.Errorf("Unexpected error: %v when uid=%d and expire=%v", err, cVals[i].uid, cVals[i].expire)
				}

				if expiresAt.Equal(cVals[i].expire) {
					t.Errorf("The expires time in the cookie didn't match the specified expiration time")
				}
			}
		})
	})
}

func TestDefaultImpl(t *testing.T) {
	auth, err := auth.New(mockJwtKey)
	if err != nil {
		t.Fatal(err)
	}

	AuthenticatorSuite(t, auth)
}
