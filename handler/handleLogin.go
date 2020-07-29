// Package handler implements all http handlers.
package handler

import (
	"auth-proxy/config"
	"auth-proxy/internal"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

// HandleLogin handles logins. If either the email or password field are invalid it returns a http.StatusBadRequest (http 400).
// If the login was successful the handler sets a language cookie and a X-CSRF header.
func HandleLogin(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body config.LoginReqBody

		err := internal.ParseJSONBody(r.Body, &body)
		if err != nil {
			http.Error(w, "Invalid request syntax", http.StatusBadRequest)
			return
		}

		if body.Email == "" || body.Pass == "" {
			http.Error(w, "Either the email or password field haven't been specified", http.StatusBadRequest)
			return
		}

		valid, err := internal.ValidateEmail(body.Email)
		if err != nil {
			http.Error(w, "An unexpected error occured. Please try again later.", http.StatusBadRequest)
			log.Println(err)
			return
		}

		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		savedUID, savedPass, lang, err := env.DB.Login(r.Context(), body)
		if err != nil {
			if err == config.ErrBadRequest {
				http.Error(w, "No user with the specified credentials exists", http.StatusBadRequest)
			} else {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
			}

			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(savedPass), []byte(body.Pass))
		if err != nil {
			http.Error(w, "You're unauthorized to perform this action", http.StatusUnauthorized)
			return
		}

		c, err := env.Auth.CreateAuthCookie(savedUID, config.DefaultExpTime())
		if err != nil {
			if err == config.ErrBadRequest {
				http.Error(w, "No user with the specified credentials exists", http.StatusBadRequest)
			} else {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
			}

			return
		}

		http.SetCookie(w, c)

		c = internal.CreateLangCookie(lang)

		http.SetCookie(w, c)

		w.Header().Set("X-CSRF-Token", csrf.Token(r))
	}
}
