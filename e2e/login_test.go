package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/errors"
)

func DoLogin(t *testing.T) {
	apiTest(t)
	if userToken != "" {
		return
	}
	assertToken := func(wantCode int, wantToken bool, saveToken bool) func(res *http.Response, req *http.Request) error {
		return func(res *http.Response, req *http.Request) error {
			if res.StatusCode != wantCode {
				return fmt.Errorf("invalid status code: %d != %d", res.StatusCode, wantCode)
			}
			data := &api.LoginResponse{}
			err := json.NewDecoder(res.Body).Decode(data)
			if err != nil {
				return fmt.Errorf("invalid json: %v", err)
			}
			if data.Token == "" && wantToken {
				return errors.New("Access token = ''")
			}
			if data.Token != "" && !wantToken {
				return errors.New("Access token != ''")
			}
			if saveToken {
				userToken = data.Token
			}
			return nil
		}
	}
	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, userName, userPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, true)).Done()
}

func DoAdminLogin(t *testing.T) {
	apiTest(t)
	if adminToken != "" {
		return
	}
	assertToken := func(wantCode int, wantToken bool, saveToken bool) func(res *http.Response, req *http.Request) error {
		return func(res *http.Response, req *http.Request) error {
			if res.StatusCode != wantCode {
				return fmt.Errorf("invalid status code: %d != %d", res.StatusCode, wantCode)
			}
			data := &api.LoginResponse{}
			err := json.NewDecoder(res.Body).Decode(data)
			if err != nil {
				return fmt.Errorf("invalid json: %v", err)
			}
			if data.Token == "" && wantToken {
				return errors.New("Access token = ''")
			}
			if data.Token != "" && !wantToken {
				return errors.New("Access token != ''")
			}
			if saveToken {
				adminToken = data.Token
			}
			return nil
		}
	}
	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, adminUser, adminPassw)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, true)).Done()
}

func TestLogin(t *testing.T) {
	apiTest(t)

	if userToken != "" {
		return
	}

	assertToken := func(wantCode int, wantToken bool, saveToken bool) func(res *http.Response, req *http.Request) error {
		return func(res *http.Response, req *http.Request) error {
			if res.StatusCode != wantCode {
				return fmt.Errorf("invalid status code: %d != %d", res.StatusCode, wantCode)
			}

			data := &api.LoginResponse{}
			err := json.NewDecoder(res.Body).Decode(data)
			if err != nil {
				return fmt.Errorf("invalid json: %v", err)
			}

			if data.Token == "" && wantToken {
				return errors.New("Access token = ''")
			}
			if data.Token != "" && !wantToken {
				return errors.New("Access token != ''")
			}

			if saveToken {
				userToken = data.Token
			}
			return nil
		}
	}

	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, userName, userPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, true)).Done()

	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, userName, "empty")).
		Expect(t).Type("json").AssertFunc(assertToken(401, false, false)).Done()

	// test that login is not case-sensitive
	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, strings.ToUpper(userName), userPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, false)).Done()
	test.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, strings.ToLower(userName), userPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, false)).Done()
}
