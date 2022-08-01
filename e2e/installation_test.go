package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

func TestServerInstallation(t *testing.T) {
	apiTest(t)
	DoAdminLogin(t)
	DoLogin(t)

	test.Authorize().client.Get("/api/v1/admin/systeminfo").Expect(t).Status(200).
		AssertFunc(func(resp *http.Response, req *http.Request) error {
			return nil
		}).Done()
	test.AuthorizeAdmin().client.Get("/api/v1/admin/systeminfo").Expect(t).Status(200).
		AssertFunc(func(resp *http.Response, req *http.Request) error {
			dto := &api.SystemInfo{}
			err := json.NewDecoder(resp.Body).Decode(dto)
			if err != nil {
				t.Errorf("parse json: %v", err)
			}

			if dto.Name != "Virtualpaper" {
				t.Errorf("expecte name 'Virtualpaper', got %s", dto.Name)
			}

			if !dto.PopplerInstalled {
				t.Errorf("poppler not installed")
			}
			if !dto.PandocInstalled {
				t.Errorf("pandoc not installed")
			}
			if !dto.SearchEngineStatus.Ok {
				t.Errorf("SearchEngineStatus not ok")
			}
			return nil
		}).Done()
}

func TestAdminGetUsers(t *testing.T) {
	apiTest(t)
	DoAdminLogin(t)
	DoLogin(t)

	test.Authorize().client.Get("/api/v1/admin/users").Expect(t).Status(401).
		AssertFunc(func(resp *http.Response, req *http.Request) error {
			return nil
		}).Done()

	test.AuthorizeAdmin().client.Get("/api/v1/admin/users").Expect(t).Status(200).
		AssertFunc(func(resp *http.Response, req *http.Request) error {
			dto := &[]models.UserInfo{}
			err := json.NewDecoder(resp.Body).Decode(dto)
			if err != nil {
				t.Errorf("parse json: %v", err)
			}
			if len(*dto) != 2 {
				t.Errorf("expect 2 users, got %d users", len(*dto))
			}
			testUser := func(user models.UserInfo, userName string, admin, active, indexing bool) {
				if user.UserName != userName {
					t.Errorf("username doesn't match: want %s, got %s", userName, user.UserName)
				}
				if user.IsAdmin != admin {
					t.Errorf("user.admin doesn't match: want %t, got %t", admin, user.IsAdmin)
				}
				if user.IsActive != active {
					t.Errorf("user.active doesn't match: want %t, got %t", active, user.IsActive)
				}
				if user.Indexing != indexing {
					t.Errorf("user.indexing doesn't match: want %t, got %t", indexing, user.Indexing)
				}
			}

			testUser((*dto)[0], "admin", true, true, false)
			testUser((*dto)[1], "user", false, true, false)
			return nil
		}).Done()

}
