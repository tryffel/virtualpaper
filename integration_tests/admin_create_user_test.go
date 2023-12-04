package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type AdminCreateUserTest struct {
	ApiTestSuite
}

func (suite *AdminCreateUserTest) SetupTest() {
	suite.Init()
	clearTestUsersTables(suite.T(), suite.db)
}

func (suite *AdminCreateUserTest) TearDownSuite() {
	clearTestUsersTables(suite.T(), suite.db)
	suite.ApiTestSuite.TearDownSuite()
}

func TestAdminCreateUser(t *testing.T) {
	suite.Run(t, new(AdminCreateUserTest))
}

func (suite *AdminCreateUserTest) TestCreateNoPermission() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid user",
		Email:         "",
		Password:      "validvalidvalid",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.userHttp, data, 401)
	assertUsersCount(&suite.ApiTestSuite, 3)
}

func (suite *AdminCreateUserTest) TestCreateBadInput() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid user",
		Email:         "",
		Password:      "short",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 400)
	assertUsersCount(&suite.ApiTestSuite, 3)

	data = &api.AdminAddUserRequest{
		UserName:      "valid user",
		Email:         "",
		Password:      "0f209a29-6db3-45c2-828c-11a81f3b3035-0f209a29-6db3-45c2-828c-11a81f3b3035-0f209a29-6db3-45c2-828c-11a81f3b3035-0f209a29-6db3-45c2-828c-11a81f3b3035f209a29-6db3-45c2-828c-11a81f3b3035-0f209a29-6db3-45c2-828c",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 400)
	assertUsersCount(&suite.ApiTestSuite, 3)

	data = &api.AdminAddUserRequest{
		UserName:      "a",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 400)
	assertUsersCount(&suite.ApiTestSuite, 3)

	data = &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "bad email",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 400)
	assertUsersCount(&suite.ApiTestSuite, 3)
}

func (suite *AdminCreateUserTest) TestCreateOk() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	assertUsersCount(&suite.ApiTestSuite, 4)

	assert.NotEqual(suite.T(), 0, user.UserId)
	assert.Equal(suite.T(), data.UserName, user.UserName)
	assert.Equal(suite.T(), data.Email, user.Email)
	assert.Equal(suite.T(), data.Active, user.IsActive)
	assert.Equal(suite.T(), data.Administrator, user.IsAdmin)

	data = &api.AdminAddUserRequest{
		UserName:      "valid name2",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user = AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	assertUsersCount(&suite.ApiTestSuite, 5)

	assert.NotEqual(suite.T(), 0, user.UserId)
	assert.Equal(suite.T(), data.UserName, user.UserName)
	assert.Equal(suite.T(), data.Email, user.Email)
	assert.Equal(suite.T(), data.Active, user.IsActive)
	assert.Equal(suite.T(), data.Administrator, user.IsAdmin)
}

func (suite *AdminCreateUserTest) TestLoginAfterCreate() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)

	token, id := LoginRequest(suite.T(), data.UserName, data.Password, 200)

	assert.Equal(suite.T(), user.UserId, id)
	assert.NotEqual(suite.T(), "", token)
}

func (suite *AdminCreateUserTest) TestLoginDisabledAfterCreate() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        false,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	LoginRequest(suite.T(), data.UserName, data.Password, 401)
}

func (suite *AdminCreateUserTest) TestCreateUserExists() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "validemail2@email.com",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	assertUsersCount(&suite.ApiTestSuite, 4)

	data = &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, http.StatusNotModified)
	assertUsersCount(&suite.ApiTestSuite, 4)

	data = &api.AdminAddUserRequest{
		UserName:      "existing user",
		Email:         "useremail@email.com",
		Password:      "utf8-ɑ-ȗ-Ǹ-ǉ",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	assertUsersCount(&suite.ApiTestSuite, 5)

	data = &api.AdminAddUserRequest{
		UserName:      "Existing USER",
		Email:         "",
		Password:      "utf8-ɑ-ȗ-Ǹ-ǉ",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 304)
	assertUsersCount(&suite.ApiTestSuite, 5)

	data = &api.AdminAddUserRequest{
		UserName:      "new user",
		Email:         "USEREMAIL@email.com",
		Password:      "utf8-ɑ-ȗ-Ǹ-ǉ",
		Active:        true,
		Administrator: false,
	}
	AdminCreateUser(suite.T(), suite.adminHttp, data, 304)
	assertUsersCount(&suite.ApiTestSuite, 5)
}

func (suite *AdminCreateUserTest) TestLoginCaseInsensitive() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	token, id := LoginRequest(suite.T(), "VALID name", data.Password, 200)
	assert.Equal(suite.T(), user.UserId, id)
	assert.NotEqual(suite.T(), "", token)
}

func AdminCreateUser(t *testing.T, client *httpClient, data *api.AdminAddUserRequest, wantHttpStatus int) *models.UserInfo {
	req := client.Post("/api/v1/admin/users")
	result := &models.UserInfo{}
	e := req.Json(t, data).ExpectName(t, "admin create user", false)
	if wantHttpStatus == 200 {
		e.Json(t, result).e.Status(200).Done()
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
	return result
}

func assertUsersCount(suite *ApiTestSuite, userCount int) {
	users, err := suite.db.UserStore.GetUsers()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), *users, userCount)
}
