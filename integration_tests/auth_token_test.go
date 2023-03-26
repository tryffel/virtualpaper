package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type AuthTokenTest struct {
	ApiTestSuite
}

func (suite *AuthTokenTest) SetupTest() {
	suite.Init()
	clearTestUsersTables(suite.T(), suite.db)
	deleteAuthTokens(&suite.ApiTestSuite)
}

func (suite *AuthTokenTest) TearDownSuite() {
	clearTestUsersTables(suite.T(), suite.db)
	deleteAuthTokens(&suite.ApiTestSuite)
	suite.ApiTestSuite.TearDownSuite()
}

func (suite *AuthTokenTest) TestLogin() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	token, _ := LoginRequest(suite.T(), data.UserName, data.Password, 200)
	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 1)
	balooClient := baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+token)
	client := &httpClient{client: balooClient}

	GetMetadataKeys(suite.T(), client, 200, nil)
	LogoutRequest(suite.T(), client, 200)
	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 0)
	// token's invalid
	LogoutRequest(suite.T(), client, 401)
	GetMetadataKeys(suite.T(), client, 401, nil)
}

func (suite *AuthTokenTest) TestFailedLogin() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	token, _ := LoginRequest(suite.T(), data.UserName, data.Password+"123", 401)
	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 0)
	balooClient := baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+token)
	client := &httpClient{client: balooClient}

	GetMetadataKeys(suite.T(), client, 401, nil)
	LogoutRequest(suite.T(), client, 401)
	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 0)
}

func (suite *AuthTokenTest) TestToken() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	LoginRequest(suite.T(), data.UserName, data.Password, 200)
	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 1)

	tokens := getUserTokens(&suite.ApiTestSuite, user.UserId)
	token := (*tokens)[0]

	assert.True(suite.T(), time.Now().Before(token.ExpiresAt), "expired yet")
	assert.NotEmpty(suite.T(), token.Key)
	assert.NotEmpty(suite.T(), token.Name)
	assert.NotEmpty(suite.T(), token.IpAddr)
}

func (suite *AuthTokenTest) TestExpire() {
	data := &api.AdminAddUserRequest{
		UserName:      "valid name",
		Email:         "",
		Password:      "passwordlongenough",
		Active:        true,
		Administrator: false,
	}
	user := AdminCreateUser(suite.T(), suite.adminHttp, data, 200)
	originalToken, _ := LoginRequest(suite.T(), data.UserName, data.Password, 200)

	tokens := getUserTokens(&suite.ApiTestSuite, user.UserId)
	token := (*tokens)[0]

	token.ExpiresAt = time.Now().Add(-time.Hour - 24)
	_, err := suite.db.Engine().Exec("update auth_tokens set expires_at = $1 where key = $2", token.ExpiresAt, token.Key)
	assert.NoError(suite.T(), err)

	assertAuthTokensCount(&suite.ApiTestSuite, user.UserId, 1)

	balooClient := baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+originalToken)
	client := &httpClient{client: balooClient}
	GetMetadataKeys(suite.T(), client, 401, nil)
}

func TestAuthTokenSuite(t *testing.T) {
	suite.Run(t, new(AuthTokenTest))
}

func assertAuthTokensCount(suite *ApiTestSuite, userId int, tokensCount int) {
	count := 0
	err := suite.db.Engine().Get(&count, "SELECT COUNT(id) FROM auth_tokens WHERE user_id = $1", userId)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), count, tokensCount)
}

func deleteAuthTokens(suite *ApiTestSuite) {
	suite.db.Engine().MustExec(`DELETE FROM auth_tokens 
       WHERE user_id NOT IN (
             SELECT id
             FROM users 
             WHERE users.name IN ('user', 'admin')
    )`)
}

func LogoutRequest(t *testing.T, client *httpClient, wantHttpStatus int) {
	client.Post("/api/v1/auth/logout").ExpectName(t, "logout", false).e.Status(wantHttpStatus).Done()
}

func getUserTokens(suite *ApiTestSuite, userId int) *[]models.Token {
	tokens := &[]models.Token{}
	err := suite.db.Engine().Select(tokens, "SELECT * FROM auth_tokens WHERE user_id=$1", userId)
	assert.NoError(suite.T(), err)
	return tokens
}
