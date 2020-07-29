package handler_test

import (
	"auth-proxy/config"
	"auth-proxy/handler"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleRegistration(t *testing.T) {
	mockEnv := config.NewMockEnv()

	cases := []struct {
		body             config.RegistrationReqBody
		expectedRespCode int
	}{
		{config.RegistrationReqBody{Email: "john.doe@gmail.de", Pass: "password", LastName: "doe", FirstName: "john"}, http.StatusOK},
		{config.RegistrationReqBody{Email: "john@doe.com", Pass: "#+kmwp/nkäwe6%hkn", LastName: "doe"}, http.StatusOK},
		{config.RegistrationReqBody{Email: "john@doe.com", Pass: "password"}, http.StatusBadRequest},
		{config.RegistrationReqBody{Email: "john@doe.com", LastName: "doe"}, http.StatusBadRequest},
		{config.RegistrationReqBody{Pass: "password", LastName: "doe"}, http.StatusBadRequest},
	}

	for _, i := range cases {
		jsonBody, err := json.Marshal(i.body)
		if err != nil {
			t.Fatal(err)
		}

		req, err := http.NewRequest("POST", "/api/register", bytes.NewReader(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handler.HandleRegistration(mockEnv))

		handler.ServeHTTP(rr, req)

		if rr.Code != i.expectedRespCode {
			t.Errorf("Expected status code %d but got %d when case=%+v", http.StatusOK, rr.Code, i.body)
		}
	}
}
