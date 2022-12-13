package integrationtest

import (
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/baloo.v3"
	"time"
)

type ApiTestSuite struct {
	suite.Suite
	userClient  *baloo.Client
	adminClient *baloo.Client

	publicHttp *httpClient
	userHttp   *httpClient
	adminHttp  *httpClient
}

func (suite *ApiTestSuite) SetupTest() {
	suite.Init()
}

func (suite *ApiTestSuite) Init() {
	EnsureUserLoggedIn(suite.T())

	suite.userClient = baloo.New(BASEURL).SetHeader("Authorization", "Bearer "+UserToken)
	suite.adminClient = baloo.New(BASEURL).SetHeader("Authorization", "Bearer "+AdminToken)

	suite.publicHttp = &httpClient{baloo.New(BASEURL)}
	suite.userHttp = &httpClient{suite.userClient}
	suite.adminHttp = &httpClient{suite.adminClient}
	clearDbMetadataTables(suite.T())
}

func isToday(date time.Time) bool {
	y1, m1, d1 := date.Date()
	y2, m2, d2 := time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
