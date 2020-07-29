package internal_test

import (
	"auth-proxy/config"
	"auth-proxy/internal"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	sptLangs := os.Getenv("SUPPORTED_LANGUAGES")

	if sptLangs == "" {
		config.SupportedLangs = []string{"en"}
	} else {
		config.SupportedLangs = strings.Split(sptLangs, ",")
	}

	os.Exit(m.Run())
}

func toJSON(t *testing.T, i interface{}) io.ReadCloser {
	t.Helper()

	b, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}

	rc := ioutil.NopCloser(bytes.NewReader(b))

	return rc
}

func TestParseJSONBody(t *testing.T) {
	var body config.LoginReqBody

	content := config.LoginReqBody{Email: "john.doe@gmail.com", Pass: "pass"}
	err := internal.ParseJSONBody(toJSON(t, content), &body)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(content, body) {
		t.Errorf("Expected the parsed body to be %v but got %v instead", content, body)
	}
}

func TestValidateEmail(t *testing.T) {
	cases := []struct {
		email string
		valid bool
	}{
		{"john.doe@gmail.com", true},
		{"john@doe.com", true},
		{"johndoe.com", false},
		{"john@doe", false},
		{"@doe.com", false},
		{"johndoe", false},
		{"john@", false},
		{"", false},
	}

	for _, i := range cases {
		emailValid, err := internal.ValidateEmail(i.email)
		if err != nil {
			t.Fatal(err)
		}

		if emailValid != i.valid {
			if i.valid {
				t.Errorf("Expected the email to be valid but it wasn't when email=%s", i.email)
			} else {
				t.Errorf("Expected the email to not be valid but it was when email=%s", i.email)
			}
		}
	}
}

func TestCreateLangCookie(t *testing.T) {
	type lang struct {
		lang  string
		valid bool
	}

	// Dynamically append supported langs to the cases array since they can change at every deploy.
	// We can still be (somewhat) sure that certain values won't be in the environment variables since
	// the specified values have to be 2 lettter ISO-639-1 codes.
	cases := []lang{
		{"wrong-lang", false},
		{"us", false},
	}
	for _, l := range config.SupportedLangs {
		cases = append(cases, lang{l, true})
	}

	for _, i := range cases {
		c := internal.CreateLangCookie(i.lang)

		if i.valid && c.Value != i.lang {
			t.Errorf("The language cookie's value (%s) didn't equal the specified value=%s", c.Value, i.lang)
		} else if !i.valid && c.Value != "en" {
			t.Errorf("The language cookie's value didn't equal \"en\" when the specified language (%s) isn't supported", i.lang)
		}
	}
}
