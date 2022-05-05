package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/errors"
)

func TestLogin(t *testing.T) {
	apiTest(t)

	assertToken := func(wantCode int, wantToken bool) func(res *http.Response, req *http.Request) error {
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
			return nil
		}
	}

	test.Post("/api/v1/auth/login").BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`,
		userName, userPassword)).SetHeader("Content-Type", "application/json").
		Expect(t).Type("json").AssertFunc(assertToken(200, true)).Done()

	test.Post("/api/v1/auth/login").BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`,
		userName, "empty")).SetHeader("Content-Type", "application/json").
		Expect(t).Type("json").AssertFunc(assertToken(401, false)).Done()

}
