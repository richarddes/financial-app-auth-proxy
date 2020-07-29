package handler

import (
	"auth-proxy/config"
	"auth-proxy/internal"
	"log"
	"net/http"

	"github.com/gorilla/csrf"
)

// HandleRegistration handles the registrations. If either the email, password or last name field is invalid
// it returns a http.StatusBadRequest (http 400).
// If successful it sets a X-CSRF header.
func HandleRegistration(env *config.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body config.RegistrationReqBody

		err := internal.ParseJSONBody(r.Body, &body)
		if err != nil {
			http.Error(w, "Invalid request syntax", http.StatusBadRequest)
			return
		}

		// firstName isn't required for registration so it doesn't have to be set
		if body.Email == "" || body.Pass == "" || body.LastName == "" {
			http.Error(w, "The email, password and last name field must have a value", http.StatusBadRequest)
			return
		}

		emailValid, err := internal.ValidateEmail(body.Email)
		if err != nil {
			http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		if !emailValid {
			http.Error(w, "Invalid request syntax", http.StatusBadRequest)
			return
		}

		err = env.DB.Register(r.Context(), body)
		if err != nil {
			if err == config.ErrBadRequest {
				http.Error(w, "An account using that email already exists", http.StatusBadRequest)
			} else {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
			}

			return
		}

		w.Header().Set("X-CSRF-Token", csrf.Token(r))
	}
}
