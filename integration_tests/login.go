package integrationtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"tryffel.net/go/virtualpaper/api"
)

const (
	UserName      = "user"
	UserPassword  = "user"
	AdminName     = "admin"
	AdminPassword = "admin"
)

var (
	UserToken  = ""
	AdminToken = ""
)

func tokenFileName() string {
	wd, _ := os.Getwd()
	fileNamee := "TOKEN"
	return path.Join(wd, fileNamee)
}

type tokenData struct {
	Comment    string
	UserToken  string
	AdminToken string
}

func ReadTokenFromFile() error {
	file := tokenFileName()
	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer fd.Close()

	data := &tokenData{}
	err = json.NewDecoder(fd).Decode(data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("parse json: %v", err)
	}
	UserToken = data.UserToken
	AdminToken = data.AdminToken
	return nil
}

func SaveTokenToFile() error {
	file := tokenFileName()

	fd, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer fd.Close()
	data := &tokenData{
		Comment:    "AUTOMATICALLY CREATED DATA. DO NOT EDIT BY HAND",
		UserToken:  UserToken,
		AdminToken: AdminToken,
	}

	err = json.NewEncoder(fd).Encode(data)
	if err != nil {
		return fmt.Errorf("write file: %v", err)
	}
	return nil
}

func DeleteTokenFile() error {
	file := tokenFileName()
	return os.Remove(file)
}

func EnsureUserLoggedIn(t *testing.T) {
	if UserToken != "" {
		return
	}

	err := ReadTokenFromFile()
	if err != nil {
		panic(fmt.Errorf("read token from file: %v", err))
	}

	if UserToken != "" {
		return
	}
	DoUserLogin(t)
	DoAdminLogin(t)
	err = SaveTokenToFile()
	if err != nil {
		panic(fmt.Errorf("save token to file: %v", err))
	}
	return
}

func DoUserLogin(t *testing.T) {
	assertToken := func(wantCode int, wantToken bool, saveToken bool) func(res *http.Response, req *http.Request) error {
		return func(res *http.Response, req *http.Request) error {
			assert.Equal(t, res.StatusCode, wantCode, "status code")
			data := &api.LoginResponse{}
			err := json.NewDecoder(res.Body).Decode(data)
			assert.Equal(t, err, nil, "invalid json", err)

			if wantToken {
				assert.NotEqual(t, data.Token, "", "token can't be empty")
			} else {
				assert.Equal(t, data.Token, "", "token must be empty")
			}
			if saveToken {
				UserToken = data.Token
			}
			return nil
		}
	}
	client.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, UserName, UserPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, true)).Done()
}

func DoAdminLogin(t *testing.T) {
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
				AdminToken = data.Token
			}
			return nil
		}
	}
	client.IsJson().client.Post("/api/v1/auth/login").
		BodyString(fmt.Sprintf(`{"Username": "%s", "Password": "%s"}`, AdminName, AdminPassword)).
		Expect(t).Type("json").AssertFunc(assertToken(200, true, true)).Done()
}
