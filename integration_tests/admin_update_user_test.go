package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type AdminUpdateUserTest struct {
	ApiTestSuite
}

func (suite *AdminUpdateUserTest) SetupTest() {
	suite.Init()
	clearTestUsersTables(suite.T(), suite.db)
}

func (suite *AdminUpdateUserTest) TearDownSuite() {
	clearTestUsersTables(suite.T(), suite.db)
	suite.ApiTestSuite.TearDownSuite()
}

func TestAdminUpdateUser(t *testing.T) {
	suite.Run(t, new(AdminUpdateUserTest))
}

func (suite *AdminUpdateUserTest) TestUpdateUserOk() {
	userData := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	originalUser := AdminCreateUser(suite.T(), suite.adminHttp, userData, 200)
	data := &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        false,
		Administrator: false,
	}

	user := AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 200)
	assert.Equal(suite.T(), user.UserName, userData.UserName)
	assert.Equal(suite.T(), user.Email, data.Email)
	assert.Equal(suite.T(), user.IsActive, data.Active)
	assert.Equal(suite.T(), user.IsAdmin, data.Administrator)
}

func (suite *AdminUpdateUserTest) TestUpdateUserFail() {
	userData := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	originalUser := AdminCreateUser(suite.T(), suite.adminHttp, userData, 200)
	data := &api.AdminUpdateUserRequest{
		Email:         "bademail",
		Password:      "",
		Active:        false,
		Administrator: false,
	}

	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 400)

	data = &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "short",
		Active:        false,
		Administrator: false,
	}
	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 400)
}

func (suite *AdminUpdateUserTest) TestAdmin() {
	userData := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	originalUser := AdminCreateUser(suite.T(), suite.adminHttp, userData, 200)

	assertUserIsActive(&suite.ApiTestSuite, originalUser.UserId, true)
	assertUserIsAdmin(&suite.ApiTestSuite, originalUser.UserId, false)

	token, _ := LoginRequest(suite.T(), userData.UserName, userData.Password, 200)

	balooClient := baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+token)
	client := &httpClient{client: balooClient}

	AdminGetUsers(suite.T(), client, 401)

	data := &api.AdminUpdateUserRequest{
		Email:         "bademail",
		Password:      "",
		Active:        false,
		Administrator: true,
	}

	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 400)
	assertUserIsAdmin(&suite.ApiTestSuite, originalUser.UserId, false)
	AdminGetUsers(suite.T(), client, 401)
	data = &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        true,
		Administrator: true,
	}
	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 200)
	assertUserIsAdmin(&suite.ApiTestSuite, originalUser.UserId, true)
	AdminGetUsers(suite.T(), client, 200)

	data = &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        true,
		Administrator: false,
	}
	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 200)
	assertUserIsAdmin(&suite.ApiTestSuite, originalUser.UserId, false)
	AdminGetUsers(suite.T(), client, 401)
}

func (suite *AdminUpdateUserTest) TestActive() {
	userData := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        false,
		Administrator: false,
	}
	originalUser := AdminCreateUser(suite.T(), suite.adminHttp, userData, 200)

	assertUserIsActive(&suite.ApiTestSuite, originalUser.UserId, false)
	assertUserIsAdmin(&suite.ApiTestSuite, originalUser.UserId, false)

	data := &api.AdminUpdateUserRequest{
		Email:         "bad email",
		Password:      "",
		Active:        false,
		Administrator: false,
	}

	LoginRequest(suite.T(), userData.UserName, userData.Password, 401)

	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 400)
	assertUserIsActive(&suite.ApiTestSuite, originalUser.UserId, false)
	data = &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        true,
		Administrator: false,
	}
	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 200)
	assertUserIsActive(&suite.ApiTestSuite, originalUser.UserId, true)

	LoginRequest(suite.T(), userData.UserName, userData.Password, 200)

	data = &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        false,
		Administrator: false,
	}
	AdminUpdateUser(suite.T(), suite.adminHttp, originalUser.UserId, data, 200)
	assertUserIsActive(&suite.ApiTestSuite, originalUser.UserId, false)

	LoginRequest(suite.T(), userData.UserName, userData.Password, 401)
}

func (suite *AdminUpdateUserTest) TestNoPermission() {
	userData := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	originalUser := AdminCreateUser(suite.T(), suite.adminHttp, userData, 200)
	data := &api.AdminUpdateUserRequest{
		Email:         "",
		Password:      "",
		Active:        false,
		Administrator: false,
	}
	AdminUpdateUser(suite.T(), suite.userHttp, originalUser.UserId, data, 401)
}

func AdminUpdateUser(t *testing.T, client *httpClient, userId int, data *api.AdminUpdateUserRequest, wantHttpStatus int) *models.UserInfo {
	req := client.Put(fmt.Sprintf("/api/v1/admin/users/%d", userId))
	result := &models.UserInfo{}
	e := req.Json(t, data).ExpectName(t, "admin update user", false)
	if wantHttpStatus == 200 {
		e.Json(t, result).e.Status(200).Done()
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
	return result
}

func assertUserIsAdmin(suite *ApiTestSuite, userId int, isAdmin bool) {
	suite.db.UserStore.FlushCache()
	user, err := suite.db.UserStore.GetUser(userId)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), isAdmin, user.IsAdmin)
}

func assertUserIsActive(suite *ApiTestSuite, userId int, isActive bool) {
	suite.db.UserStore.FlushCache()
	user, err := suite.db.UserStore.GetUser(userId)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), isActive, user.IsActive)
}

func AdminGetUsers(t *testing.T, client *httpClient, wantHttpStatus int) {
	client.Get("/api/v1/admin/users").Expect(t).e.Status(wantHttpStatus).Done()
}
