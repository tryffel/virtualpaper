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
	// document can still be viewed while it's in the trashbin
	getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
}

func (suite *DocumentDeleteSuite) TestRestoreDocument() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 60)
	filter := map[string]string{
		"q": "Vestibulum sit amet dignissim nun",
	}
	waitIndexingReady(suite.T(), suite.userHttp, 60)

	originalHistory := getDocumentHistory(suite.T(), suite.userHttp, docId, 200)
	requestDeleteDocument(suite.T(), suite.userHttp, docId, 200)
	assertDocumentHasDeletetAt(&suite.ApiTestSuite, docId, time.Now())

	history := getDocumentHistory(suite.T(), suite.userHttp, docId, 200)
	assert.Len(suite.T(), *history, len(*originalHistory)+1)
	assert.Equal(suite.T(), (*history)[len(*history)-1].Action, models.DocumentHistoryActionDelete)

	waitIndexingReady(suite.T(), suite.userHttp, 60)
	docs := searchDocuments(suite.T(), suite.userHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 0, len(docs))

	requestRestoreDeletedDocument(suite.T(), suite.userHttp, docId, 200)
	assertDocumentHasNoDeletedAt(&suite.ApiTestSuite, docId)

	requestRestoreDeletedDocument(suite.T(), suite.userHttp, docId, 404)

	waitIndexingReady(suite.T(), suite.userHttp, 60)
	docs = searchDocuments(suite.T(), suite.userHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 1, len(docs))

	history = getDocumentHistory(suite.T(), suite.userHttp, docId, 200)
	assert.Len(suite.T(), *history, len(*originalHistory)+2)
	assert.Equal(suite.T(), (*history)[len(*history)-1].Action, models.DocumentHistoryActionRestore)
}

func (suite *DocumentDeleteSuite) TestFlushDeletedDocument() {
	insertTestDocuments(suite.T(), suite.db)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assertDocumentHasDeletetAt(&suite.ApiTestSuite, testDocumentX86Intel.Id, time.Now())

	getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	requestFlushDeletedDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 404)
}

func (suite *DocumentDeleteSuite) TestDeleteInvalid() {
	insertTestDocuments(suite.T(), suite.db)

	// cannot delete other user's document
	requestDeleteDocument(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 404)

	// deleting document twice fails
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 400)
	assertDocumentHasDeletetAt(&suite.ApiTestSuite, testDocumentX86Intel.Id, time.Now())

	// cannot restore other user's document
	requestRestoreDeletedDocument(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 404)
	requestFlushDeletedDocument(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 404)

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

func (suite *DocumentDeleteSuite) TestListDocuments() {
	insertTestDocuments(suite.T(), suite.db)
	totalDocs := len(testDocuments) - 1

	docs := getDocuments(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), *docs, totalDocs)

	deletedDocs := getDeletedDocuments(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), *deletedDocs, 0)

	assertDocumentInArray(suite.T(), testDocumentX86Intel.Id, docs)
	assertDocumentInArray(suite.T(), testDocumentJupiterMoons.Id, docs)

	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentJupiterMoons.Id, 200)

	docs = getDocuments(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), *docs, totalDocs-2)

	deletedDocs = getDeletedDocuments(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), *deletedDocs, 2)

	assertDocumentInArray(suite.T(), testDocumentX86Intel.Id, deletedDocs)
	assertDocumentInArray(suite.T(), testDocumentJupiterMoons.Id, deletedDocs)
	assertDocumentNotInArray(suite.T(), testDocumentX86Intel.Id, docs)
	assertDocumentNotInArray(suite.T(), testDocumentJupiterMoons.Id, docs)

	requestRestoreDeletedDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docs = getDocuments(suite.T(), suite.userHttp, 200)
	deletedDocs = getDeletedDocuments(suite.T(), suite.userHttp, 200)
	assertDocumentInArray(suite.T(), testDocumentX86Intel.Id, docs)
	assertDocumentNotInArray(suite.T(), testDocumentX86Intel.Id, deletedDocs)
}

func (suite *DocumentDeleteSuite) TestDocumentStatistics() {
	insertTestDocuments(suite.T(), suite.db)

	getDocumentWithVisit(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	getDocumentWithVisit(suite.T(), suite.userHttp, testDocumentJupiterMoons.Id, 200)
	stats := getDocumentStatistics(suite.T(), suite.userHttp, 200)

	// one of the documents is another user's
	originalLength := len(testDocuments) - 1
	assert.Len(suite.T(), stats.LastDocumentsUpdated, originalLength)
	assert.Len(suite.T(), stats.LastDocumentsAdded, originalLength)
	assert.Len(suite.T(), stats.LastDocumentsViewed, 2)

	requestDeleteDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	stats = getDocumentStatistics(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), stats.LastDocumentsUpdated, originalLength-1)
	assert.Len(suite.T(), stats.LastDocumentsAdded, originalLength-1)
	assert.Len(suite.T(), stats.LastDocumentsViewed, 1)

	requestRestoreDeletedDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	stats = getDocumentStatistics(suite.T(), suite.userHttp, 200)
	assert.Len(suite.T(), stats.LastDocumentsUpdated, originalLength)
	assert.Len(suite.T(), stats.LastDocumentsAdded, originalLength)
	assert.Len(suite.T(), stats.LastDocumentsViewed, 2)
}

func requestDeleteDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Delete(fmt.Sprintf("/api/v1/documents/%s", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func requestFlushDeletedDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Delete(fmt.Sprintf("/api/v1/documents/deleted/%s", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func requestRestoreDeletedDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Post(fmt.Sprintf("/api/v1/documents/deleted/%s/restore", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func requestAdminRestoreDeletedDocument(t *testing.T, client *httpClient, docId string, wantHttpStatus int) {
	client.Post(fmt.Sprintf("/api/v1/admin/documents/deleted/%s/restore", docId)).Expect(t).e.Status(wantHttpStatus).Done()
}

func assertDocumentHasDeletetAt(suite *ApiTestSuite, docId string, date time.Time) {
	doc, err := suite.db.DocumentStore.GetDocument(suite.db, docId)
	assert.NoError(suite.T(), err)

	midnight := models.MidnightForDate(date)
	assert.True(suite.T(), doc.DeletedAt.Valid)
	docMidnight := models.MidnightForDate(doc.DeletedAt.Time)

	assert.Equal(suite.T(), midnight, docMidnight)
}

func assertDocumentHasNoDeletedAt(suite *ApiTestSuite, docId string) {
	doc, err := suite.db.DocumentStore.GetDocument(suite.db, docId)
	assert.NoError(suite.T(), err)

	assert.False(suite.T(), doc.DeletedAt.Valid)
	assert.True(suite.T(), doc.DeletedAt.Time.IsZero())
}
