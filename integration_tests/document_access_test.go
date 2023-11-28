package integrationtest

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type DocumentAccessTest struct {
	ApiTestSuite
}

func (suite *DocumentAccessTest) SetupTest() {
	suite.Init()
	insertTestDocuments(suite.T(), suite.db)
}

func (suite *DocumentAccessTest) TearDownSuite() {
	suite.ApiTestSuite.TearDownSuite()
}

func TestDocumentAccess(t *testing.T) {
	suite.Run(t, new(DocumentAccessTest))
}

func (suite *DocumentAccessTest) TestDownloadForbidden() {
	doc1 := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60)
	doc2 := uploadDocument(suite.T(), suite.adminClient, "text-1.txt", "Lorem ipsum", 60)

	downloadDocument(suite.T(), suite.userClient, doc1, 200)
	downloadDocument(suite.T(), suite.adminClient, doc1, 404)

	downloadDocument(suite.T(), suite.adminClient, doc2, 200)
	downloadDocument(suite.T(), suite.userClient, doc2, 404)
}

func (suite *DocumentAccessTest) TestShowForbidden() {
	getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	getDocument(suite.T(), suite.adminHttp, testDocumentMetamorphosis.Id, 404)

	getDocument(suite.T(), suite.adminHttp, testDocumentTransistorCountAdminUser.Id, 200)
	getDocument(suite.T(), suite.userHttp, testDocumentTransistorCountAdminUser.Id, 404)
}
