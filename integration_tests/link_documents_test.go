package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"tryffel.net/go/virtualpaper/models"
)

func TestDocumentLinking(t *testing.T) {
	suite.Run(t, new(DocumentLinkingTestSuite))
}

type DocumentLinkingTestSuite struct {
	ApiTestSuite
}

func (suite *DocumentLinkingTestSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T(), suite.db)
	_ = insertTestDocuments(suite.T(), suite.db)
}

func (suite *DocumentLinkingTestSuite) TestCreateLink() {
	docs := getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 0)
	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, []string{testDocumentX86.Id}, 200)
	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 1)
	assert.Equal(suite.T(), testDocumentX86.Id, (*docs)[0].DocumentId)

	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *docs, 1)
	assert.Equal(suite.T(), testDocumentX86Intel.Id, (*docs)[0].DocumentId)
}

func (suite *DocumentLinkingTestSuite) TestCreateSeveralLinks() {
	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, []string{testDocumentX86.Id, testDocumentMetamorphosis.Id}, 200)
	docs := getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 2)

	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *docs, 1)
	assert.Equal(suite.T(), testDocumentX86Intel.Id, (*docs)[0].DocumentId)

	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Len(suite.T(), *docs, 1)
	assert.Equal(suite.T(), testDocumentX86Intel.Id, (*docs)[0].DocumentId)
}

func (suite *DocumentLinkingTestSuite) TestDeleteLink() {
	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, []string{testDocumentX86.Id, testDocumentMetamorphosis.Id}, 200)
	docs := getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 2)

	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, []string{}, 200)
	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Len(suite.T(), *docs, 0)

	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 1)

	docs = getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *docs, 1)
}

func (suite *DocumentLinkingTestSuite) TestInvalidPermissions() {
	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, []string{testDocumentTransistorCountAdminUser.Id}, 400)
	docs := getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 0)
}

func (suite *DocumentLinkingTestSuite) TestDocumentNotFound() {
	updateLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, []string{"1234", testDocumentX86.Id}, 400)
	updateLinkedDocuments(suite.T(), suite.userHttp, "1234", []string{testDocumentX86.Id}, 404)
	docs := getLinkedDocuments(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Len(suite.T(), *docs, 0)
}

func getLinkedDocuments(t *testing.T, client *httpClient, docId string, wantHttpStatus int) *[]models.LinkedDocument {
	data := &[]models.LinkedDocument{}
	req := client.Get(fmt.Sprintf("/api/v1/documents/%s/linked-documents", docId))
	req.ExpectName(t, "get linked documents", false).Json(t, data).e.Status(wantHttpStatus).Done()
	return data
}

func updateLinkedDocuments(t *testing.T, client *httpClient, docId string, docs []string, wantHttpStatus int) {
	data := map[string]interface{}{"documents": docs}
	req := client.Put(fmt.Sprintf("/api/v1/documents/%s/linked-documents", docId)).Json(t, data)
	req.ExpectName(t, "", false).e.Status(wantHttpStatus).Done()
}
