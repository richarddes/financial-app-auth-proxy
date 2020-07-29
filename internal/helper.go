// Package internal provides common globally used functions.
package internal

import (
	"auth-proxy/config"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"time"
)

// ParseJSONBody decodes a jsonBody and saves it in the value pointed by v.
func ParseJSONBody(jsonBody io.ReadCloser, v interface{}) error {
	err := json.NewDecoder(jsonBody).Decode(v)
	if err != nil {
		return err
	}

	return nil
}

// ValidateEmail checks if the specified email matches a rough email pattern.
func ValidateEmail(email string) (bool, error) {
	emailRegexMatched, err := regexp.MatchString(`^.+@\w+\.\w+$`, email)
	if err != nil {
		return false, err
	}

	if !emailRegexMatched {
		return false, nil
	}

	return true, nil
}

// CreateLangCookie creates a language cookie with the specified lang paraeter as it's value.
// If the specified value isn't a value specified in the config's SupportedLangs array it sets
// "en" (English) as the default value.
func CreateLangCookie(lang string) *http.Cookie {
	if !IsSupportedLang(lang) {
		lang = "en"
	}

	monthDuration := time.Now().Add(time.Hour * 5040) // actually 30 days not a month

	c := &http.Cookie{
		Name:     "lang",
		Path:     "/",
		Value:    lang,
		Expires:  monthDuration,
		SameSite: http.SameSiteStrictMode,
	}

	return c
}

// IsSupportedLang checks if the specified language is in the config's SupportedLangs array.
func IsSupportedLang(lang string) bool {
	for _, i := range config.SupportedLangs {
		if lang == i {
			return true
		}
	}

	return false
}
