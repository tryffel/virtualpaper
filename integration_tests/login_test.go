package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/suite"
	"log"
	"testing"
)

type LoginTestSuite struct {
	ApiTestSuite
}

func (suite *LoginTestSuite) SetupTest() {
	err := DeleteTokenFile()
	if err != nil {
		panic(fmt.Errorf("delete token file: %v", err))
	}
}

func (suite *LoginTestSuite) TestLogin() {
	log.Println(suite.T().Name())
	DoUserLogin(suite.T())
	DoAdminLogin(suite.T())
	DoTesterLogin(suite.T())
	EnsureUserLoggedIn(suite.T())
	err := SaveTokenToFile()
	if err != nil {
		panic(fmt.Errorf("save token to file: %v", err))
	}
}

func TestLoginSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}
