package integrationtest

import (
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/storage"
)

type ApiTestSuite struct {
	suite.Suite
	userClient  *baloo.Client
	adminClient *baloo.Client

	publicHttp *httpClient
	userHttp   *httpClient
	adminHttp  *httpClient

	db *storage.Database
}

func (suite *ApiTestSuite) SetupSuite() {
	suite.db = GetDb()
}

func (suite *ApiTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *ApiTestSuite) SetupTest() {
	if testing.Short() {
		suite.T().SkipNow()
	}

	initConfig()
	suite.Init()
}

func (suite *ApiTestSuite) Init() {
	EnsureUserLoggedIn(suite.T())

	suite.userClient = baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+UserToken)
	suite.adminClient = baloo.New(serverUrl).SetHeader("Authorization", "Bearer "+AdminToken)

	suite.publicHttp = &httpClient{baloo.New(serverUrl)}
	suite.userHttp = &httpClient{suite.userClient}
	suite.adminHttp = &httpClient{suite.adminClient}
	clearDbMetadataTables(suite.T(), suite.db)
}

func isToday(date time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func init() {
	initConfig()
}
