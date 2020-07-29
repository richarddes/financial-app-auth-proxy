package handler_test

import (
	"auth-proxy/auth"
	"auth-proxy/config"
	"auth-proxy/handler"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleLogin(t *testing.T) {
	auth.New("jwt-key")

	ctx := context.Background()
	mockEnv := config.NewMockEnv()

	bodies := []config.RegistrationReqBody{
		{Email: "john.doe@gmail.com", Pass: "password", LastName: "doe", FirstName: "john"},
		{Email: "john@doe.com", Pass: "#+kmwp/nkäwe6%hkn", LastName: "doe"},
	}

	for _, b := range bodies {
		err := mockEnv.DB.Register(ctx, b)
		if err != nil {
			log.Fatal(err)
		}
	}

	cases := []struct {
		body         config.LoginReqBody
		expectedCode int
	}{
		{config.LoginReqBody{Email: "john.doe@gmail.com", Pass: "password"}, http.StatusOK},
		{config.LoginReqBody{Email: "john@doe.com", Pass: "#+kmwp/nkäwe6%hkn"}, http.StatusOK},
		{config.LoginReqBody{Email: "john@doe.com"}, http.StatusBadRequest},
		{config.LoginReqBody{Pass: "password"}, http.StatusBadRequest},
		{config.LoginReqBody{}, http.StatusBadRequest},
	}

	for _, i := range cases {
		jsonBody, err := json.Marshal(i.body)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/api/login", bytes.NewReader(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handler.HandleLogin(mockEnv))

		handler.ServeHTTP(rr, req)

		if rr.Code != i.expectedCode {
			t.Errorf("Expected status code %d but got %d when case=%v", i.expectedCode, rr.Code, i.body)
		}

		n := len(rr.Result().Cookies())

		if i.expectedCode == http.StatusOK {
			if n != 2 {
				t.Errorf("Expected an authentication and language cookie to be set but got %v when body=%v", rr.Result().Cookies(), i.body)
			}
		} else {
			if n > 0 {
				t.Errorf("Expected no cookies to be set but got %v when body=%v", rr.Result().Cookies(), i.body)
			}
		}
	}
}
