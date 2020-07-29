// +build integration

package models_test

import (
	"auth-proxy/config"
	"auth-proxy/models"
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASSWORD")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbHost = os.Getenv("DB_HOST")
)

func init() {
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

	// test values
	config.SupportedLangs = []string{"en", "de"}
}

func TestDefaulImpl(t *testing.T) {
	connStr := fmt.Sprintf("port=%s user=%s password=%s dbname=%s host=%s sslmode=disable", dbPort, dbUser, dbPass, dbName, dbHost)
	db, err := models.New(connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	ModelsSuite(t, db)
}

func ModelsSuite(t *testing.T, impl config.Datastore) {
	ctx := context.Background()

	cases := []struct {
		body  config.RegistrationReqBody
		valid bool
	}{
		{config.RegistrationReqBody{Email: "john@doe.com", FirstName: "john", LastName: "doe"}, false},
		{config.RegistrationReqBody{Pass: "password", FirstName: "john"}, false},
		{config.RegistrationReqBody{}, false},
		{config.RegistrationReqBody{Email: "john.doe@gmail.de", Pass: "password", FirstName: "john", LastName: "doe"}, true},
		{config.RegistrationReqBody{Email: "john@doe.com", Pass: "#+kmwp/nkdwäe6%hkn", LastName: "doe"}, true},
		{config.RegistrationReqBody{Email: "john@doe.com", Pass: "password", LastName: "doe"}, true},
	}

	t.Run("Testing registration handler", func(t *testing.T) {
		for _, i := range cases {
			err := impl.Register(ctx, i.body)

			if i.valid {
				if err != nil {
					t.Fatalf("Unexpected error: %v when body=%v", err, i.body)
				}
			} else {
				if err == nil {
					t.Errorf("Expected err but got nil when body=%v", i.body)
				}
			}
		}
	})

	t.Run("Testing login handler", func(t *testing.T) {
		loginCases := []struct {
			body  config.LoginReqBody
			valid bool
		}{
			{config.LoginReqBody{Email: "john@doe.com"}, false},
			{config.LoginReqBody{Pass: "password"}, false},
			{config.LoginReqBody{}, false},
			{config.LoginReqBody{Email: "john@doe.de", Pass: "#+kmwp/nkdwäe6%hkn"}, true},
			{config.LoginReqBody{Email: "john.doe@gmail.com", Pass: "password"}, true},
		}

		for _, i := range loginCases {
			uid, pass, _, err := impl.Login(ctx, i.body)

			if i.valid {
				if err != nil {
					t.Fatalf("Unexpected error %v when body=%v", err, i.body)
				}

				if uid < 0 {
					t.Errorf("User-id was smaller than 0 when body=%v", i.body)
				}
			} else {
				if err == nil {
					t.Errorf("Expected err but got nil when body=%v", i.body)
				}

				if pass != "" {
					t.Errorf("A password was returned when the body=%v", i.body)
				}
			}
		}
	})
}
