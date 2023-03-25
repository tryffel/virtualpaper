package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

type PasswordReset struct {
	ApiTestSuite
	user *models.User
}

func (suite *PasswordReset) SetupTest() {
	suite.Init()
	clearPasswordResetTables(suite.T(), suite.db)

	user, err := suite.db.UserStore.GetUserByName("user")
	if err != nil {
		suite.T().Errorf("get user by name: %v", err)
		return
	}
	suite.user = user
	user.Email = "testperson@test.com"
	err = suite.db.UserStore.Update(user)
	if err != nil {
		suite.T().Errorf("update user email: %v", err)
		return
	}
}

func TestGetPasswordResetLink(t *testing.T) {
	suite.Run(t, new(PasswordReset))
}

func (suite *PasswordReset) TestEmailNotExists() {
	RequestResetPasswordToken(suite.T(), suite.userHttp, "notexists@test.com", 200)
	tokens, err := GetPasswordResetTokensFromDb(suite.db)
	assert.Nil(suite.T(), err)
	assert.Equal(suite.T(), 0, len(*tokens), "no reset tokens found")
}

func (suite *PasswordReset) TestEmailExists() {
	RequestResetPasswordToken(suite.T(), suite.userHttp, "testperson@test.com", 200)
	tokens, err := GetPasswordResetTokensFromDb(suite.db)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), *tokens, 1, "one reset token found")

	token := (*tokens)[0]
	today := models.MidnightForDate(time.Now())
	tomorrow := models.MidnightForDate(time.Now().Add(time.Hour * 24))
	assert.True(suite.T(), len(token.Token) > 40, "token length at least 40 characters")
	assert.Equal(suite.T(), suite.user.Id, token.UserId)

	createdAt := models.MidnightForDate(token.CreatedAt)
	expiresAt := models.MidnightForDate(token.ExpiresAt)

	assert.True(suite.T(), today.Equal(createdAt), "token created today")
	assert.True(suite.T(), tomorrow.Equal(expiresAt), "token expired next day")
}

func (suite *PasswordReset) TestTokensAreUnique() {
	RequestResetPasswordToken(suite.T(), suite.userHttp, "testperson@test.com", 200)
	RequestResetPasswordToken(suite.T(), suite.userHttp, "testperson@test.com", 200)
	RequestResetPasswordToken(suite.T(), suite.userHttp, "testperson@test.com", 200)
	tokens, err := GetPasswordResetTokensFromDb(suite.db)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), *tokens, 3, "three reset tokens found")

	tokenMap := map[string]bool{}

	for _, v := range *tokens {
		tokenMap[v.Token] = true
	}
	assert.Len(suite.T(), tokenMap, 3)
}

/* TODO: need to retrieve token from the email
func (suite *PasswordReset) TestResetPasswordWithToken() {
	RequestResetPasswordToken(suite.T(), suite.userHttp, "testperson@test.com", 200)
	tokens, err := GetPasswordResetTokensFromDb(suite.db)
	assert.Nil(suite.T(), err)
	assert.Len(suite.T(), *tokens, 1, "one reset token found")

	token := (*tokens)[0]

	ResetPassword(suite.T(), suite.userHttp, token.Token, "new-password123", token.Id, 200)

	newToken, userId := LoginRequest(suite.T(), "user", "new-password123", 200)
	assert.NotNil(suite.T(), newToken)
	assert.Equal(suite.T(), suite.user.Id, userId)
}
*/

func RequestResetPasswordToken(t *testing.T, client *httpClient, email string, wantHttpStatus int) {
	req := client.Post("/api/v1/auth/forgot-password")
	dto := &api.ForgottenPasswordRequest{Email: email}

	body := map[string]string{}

	e := req.Json(t, dto).ExpectName(t, "get password reset token", false)
	if wantHttpStatus == 200 {
		e.Json(t, &body).e.Status(200).Done()
		msg := body["status"]
		assert.Equal(t, "Password reset email has been sent", msg)
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
}

func GetPasswordResetTokensFromDb(db *storage.Database) (*[]models.PasswordResetToken, error) {
	sql := "SELECT * FROM password_reset_tokens"
	data := &[]models.PasswordResetToken{}
	err := db.Engine().Select(data, sql)
	return data, err
}

func ResetPassword(t *testing.T, client *httpClient, token, password string, id int, wantHttpStatus int) {
	req := client.Post("/api/v1/auth/reset-password")
	dto := &api.ResetPasswordRequest{
		Token:    token,
		Id:       id,
		Password: password,
	}
	body := map[string]string{}
	e := req.Json(t, dto).ExpectName(t, "reset password", true)
	if wantHttpStatus == 200 {
		e.Json(t, &body).e.Status(200).Done()
		msg := body["status"]
		assert.Equal(t, "Password reset email has been sent", msg)
	} else {
		e.e.Status(wantHttpStatus).Done()
	}
}
