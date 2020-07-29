package main

import (
	"auth-proxy/auth"
	"auth-proxy/config"
	"auth-proxy/handler"
	"auth-proxy/internal"
	"auth-proxy/models"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

var (
	jwtKey   = os.Getenv("JWT_KEY")
	csrfKey  = os.Getenv("CSRF_KEY")
	dbUser   = os.Getenv("DB_USER")
	dbPass   = os.Getenv("DB_PASSWORD")
	dbPort   = os.Getenv("DB_PORT")
	dbName   = os.Getenv("DB_NAME")
	dbHost   = os.Getenv("DB_HOST")
	sptLangs = os.Getenv("SUPPORTED_LANGUAGES")

	env *config.Env
)

func init() {
	if jwtKey == "" {
		log.Fatal("No environment variable named JWT_KEY present")
	}

	if csrfKey == "" {
		log.Fatal("No environment variable named CSRF_KEY present")
	}

	if dbUser == "" {
		log.Fatal("No environment variable named DB_USER present")
	}

	if dbPass == "" {
		log.Fatal("No environment variable named DB_PASSWORD present")
	}

	if dbPort == "" {
		log.Fatal("No environment variable named DB_PORT present")
	}

	if dbName == "" {
		log.Fatal("No environment variable named DB_NAME present")
	}

	if dbHost == "" {
		dbHost = "localhost"
	}

	if sptLangs == "" {
		config.SupportedLangs = []string{"en"}
	}

	config.SupportedLangs = strings.Split(sptLangs, ",")
}

func main() {
	connStr := fmt.Sprintf("port=%s user=%s password=%s dbname=%s host=%s sslmode=disable", dbPort, dbUser, dbPass, dbName, dbHost)
	db, err := models.New(connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	auth, err := auth.New(jwtKey)
	if err != nil {
		log.Fatal(err)
	}

	env = &config.Env{DB: db, Auth: auth}

	usersProxy, err := routeReverseProxy("user_service:8081", "user_service")
	if err != nil {
		log.Panic(err)
	}

	newsProxy, err := routeReverseProxy("news_service:8083", "news_service")
	if err != nil {
		log.Panic(err)
	}

	stockProxy, err := routeReverseProxy("stock_service:8082", "stock_service")
	if err != nil {
		log.Panic(err)
	}

	csrfMiddleware := csrf.Protect(
		[]byte(csrfKey),
	)

	r := mux.NewRouter()
	r.Use(csrfMiddleware)

	api := r.PathPrefix("/api").Subrouter()

	api.HandleFunc("/get-csrf-token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
	}).Methods("GET")

	api.Handle("/check-credentials", authMiddleware(func() http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {}
	}())).Methods("GET")

	api.Handle("/login", handler.HandleLogin(env)).Methods("POST")
	api.Handle("/register", handler.HandleRegistration(env)).Methods("POST")

	apiUsers := api.PathPrefix("/users").Subrouter()
	apiUsers.Use(authMiddleware)
	apiUsers.NewRoute().Handler(usersProxy)

	apiNews := api.PathPrefix("/news").Subrouter()
	apiNews.Use(authMiddleware)
	apiNews.NewRoute().Handler(newsProxy)

	apiStocks := api.PathPrefix("/stocks").Subrouter()
	apiStocks.Use(authMiddleware)
	apiStocks.NewRoute().Handler(stockProxy)

	fmt.Println("The auth proxy is ready")
	log.Panic(http.ListenAndServe(":9000", r))
}

func routeReverseProxy(host, servicePath string) (*httputil.ReverseProxy, error) {
	serviceURL, err := url.Parse(servicePath)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(serviceURL)

	origin, err := url.Parse("http_server")
	if err != nil {
		return nil, err
	}

	director := func(r *http.Request) {
		r.Header.Add("X-Forwarded-Host", r.Host)
		r.Header.Add("X-Origin-Host", origin.Host)
		r.URL.Scheme = "http"
		r.URL.Host = host
	}
	proxy.Director = director

	modifyResp := func(resp *http.Response) error {
		r := resp.Request

		if r.Method == "POST" {
			resp.Header.Set("X-CSRF-Token", csrf.Token(r))
		}

		return nil
	}
	proxy.ModifyResponse = modifyResp

	return proxy, nil
}

func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("auth_token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "The request didn't include an authentication token", http.StatusUnauthorized)
			} else {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
			}

			return
		}

		err = env.Auth.ValidateAuthCookie(c)
		if err != nil {
			http.Error(w, "The specified authentication token's invalid", http.StatusBadRequest)
			return
		}

		uid, err := env.Auth.UID(c)
		if err != nil {
			http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		expiresAt, err := env.Auth.ExpiresAt(c)
		if err != nil {
			http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
			log.Println(err)
			return
		}

		// refresh the token if it's about to expire
		if time.Until(expiresAt) < time.Second*30 {
			c, err = env.Auth.CreateAuthCookie(uid, config.DefaultExpTime())
			if err != nil {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
				return
			}

			http.SetCookie(w, c)
		}

		c, err = r.Cookie("lang")
		if err != nil {
			if err == http.ErrNoCookie {
				c = internal.CreateLangCookie("en")

				http.SetCookie(w, c)
			} else {
				http.Error(w, "An unexpected error occured. Please try again later.", http.StatusInternalServerError)
				log.Println(err)
				return
			}
		}

		lang := c.Value

		if !internal.IsSupportedLang(lang) {
			http.Error(w, "The specified language isn't a supported language", http.StatusBadRequest)
			return
		}

		r.Header.Add("UID", strconv.FormatUint(uid, 10))
		r.Header.Add("Lang", lang)

		h.ServeHTTP(w, r)
	})
}
