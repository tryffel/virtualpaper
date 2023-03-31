package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func TestDeleteDocument(t *testing.T) {
	suite.Run(t, new(DocumentDeleteSuite))
}

type DocumentDeleteSuite struct {
	ApiTestSuite
}

func (suite *DocumentDeleteSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T(), suite.db)
	clearMeiliIndices(suite.T())
}

func (suite *DocumentDeleteSuite) TestDeleteDocument() {
	insertTestDocuments(suite.T(), suite.db)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	assertDocumentHasDeletetAt(&suite.ApiTestSuite, testDocumentX86Intel.Id, time.Now())
}

func (suite *DocumentDeleteSuite) TestDeleteInvalid() {
	insertTestDocuments(suite.T(), suite.db)

	// cannot delete other user's document
	requestDeleteDocument(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 403)

	// deleting document twice succeeds
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assertDocumentHasDeletetAt(&suite.ApiTestSuite, testDocumentX86Intel.Id, time.Now())
}

func (suite *DocumentDeleteSuite) TestAdminRestoreDeletedDocument() {
	insertTestDocuments(suite.T(), suite.db)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assertDocumentHasDeletetAt(&suite.ApiTestSuite, testDocumentX86Intel.Id, time.Now())

	requestAdminRestoreDeletedDocument(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 200)
	assertDocumentHasNoDeletedAt(&suite.ApiTestSuite, testDocumentX86Intel.Id)
}

func (suite *DocumentDeleteSuite) TestSearchingDeletedDocumentFails() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60)

	filter := map[string]string{
		"q": "Vestibulum sit amet dignissim nun",
	}
	waitIndexingReady(suite.T(), suite.userHttp, 60)

	docs := searchDocuments(suite.T(), suite.userHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 1, len(docs))
	assert.Equal(suite.T(), docId, docs[0].Id)

	requestDeleteDocument(suite.T(), suite.userHttp, docId, 200)

	waitIndexingReady(suite.T(), suite.userHttp, 60)
	docs = searchDocuments(suite.T(), suite.userHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 0, len(docs))

	requestAdminRestoreDeletedDocument(suite.T(), suite.adminHttp, docId, 200)
	assertDocumentHasNoDeletedAt(&suite.ApiTestSuite, docId)
	waitIndexingReady(suite.T(), suite.userHttp, 60)

	docs = searchDocuments(suite.T(), suite.userHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 1, len(docs))
	assert.Equal(suite.T(), docId, docs[0].Id)
}

func requestDeleteDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Delete(fmt.Sprintf("/api/v1/documents/%s", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func requestAdminRestoreDeletedDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Post(fmt.Sprintf("/api/v1/admin/documents/deleted/%s/restore", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func assertDocumentHasDeletetAt(suite *ApiTestSuite, docId string, date time.Time) {
	doc, err := suite.db.DocumentStore.GetDocument(0, docId)
	assert.NoError(suite.T(), err)

	midnight := models.MidnightForDate(date)
	assert.True(suite.T(), doc.DeletedAt.Valid)
	docMidnight := models.MidnightForDate(doc.DeletedAt.Time)

	assert.Equal(suite.T(), midnight, docMidnight)
}

func assertDocumentHasNoDeletedAt(suite *ApiTestSuite, docId string) {
	doc, err := suite.db.DocumentStore.GetDocument(0, docId)
	assert.NoError(suite.T(), err)

	assert.False(suite.T(), doc.DeletedAt.Valid)
	assert.True(suite.T(), doc.DeletedAt.Time.IsZero())
}
