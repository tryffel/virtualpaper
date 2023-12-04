package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

func TestDocumentSharing(t *testing.T) {
	suite.Run(t, new(ShareDocumentTestSuite))
}

type ShareDocumentTestSuite struct {
	ApiTestSuite
	users map[string]models.User
}

func (suite *ShareDocumentTestSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T(), suite.db)
	_ = insertTestDocuments(suite.T(), suite.db)
	clearMeiliIndices(suite.T())
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	users, err := suite.db.UserStore.GetUsers()
	if err != nil {
		suite.T().Error("get users from db", err)
	} else {
		suite.users = map[string]models.User{}
		for _, v := range *users {
			suite.users[v.Name] = v
		}
	}
}

func (suite *ShareDocumentTestSuite) TestShareDocument() {
	userDocs := getDocuments(suite.T(), suite.userHttp, 200)
	adminDocs := getDocuments(suite.T(), suite.adminHttp, 200)
	testerDocs := getDocuments(suite.T(), suite.testerHttp, 200)
	assert.Equal(suite.T(), 6, len(*userDocs), "no documents shared")
	assert.Equal(suite.T(), 1, len(*adminDocs), "no documents shared")
	assert.Equal(suite.T(), 0, len(*testerDocs), "no documents shared")

	doc := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["admin"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  true,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.adminHttp, doc.Id, request, 404)
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)

	userDocs = getDocuments(suite.T(), suite.userHttp, 200)
	adminDocs = getDocuments(suite.T(), suite.adminHttp, 200)
	testerDocs = getDocuments(suite.T(), suite.testerHttp, 200)
	assert.Equal(suite.T(), 6, len(*userDocs), "no documents shared")
	assert.Equal(suite.T(), 2, len(*adminDocs), "no documents shared")
	assert.Equal(suite.T(), 1, len(*testerDocs), "no documents shared")
}

func (suite *ShareDocumentTestSuite) TestUnshareDocument() {
	doc := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Equal(suite.T(), 0, doc.Shares)
	assert.Empty(suite.T(), doc.SharedUsers)
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["admin"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  true,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.adminHttp, doc.Id, request, 404)
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)

	doc = getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Equal(suite.T(), 2, doc.Shares)
	assert.NotEmpty(suite.T(), doc.SharedUsers)
	assert.Equal(suite.T(), doc.SharedUsers[0].UserId, suite.users["admin"].Id)
	assert.Equal(suite.T(), true, doc.SharedUsers[0].Permissions.Read)
	assert.Equal(suite.T(), false, doc.SharedUsers[0].Permissions.Write)
	assert.Equal(suite.T(), doc.SharedUsers[1].UserId, suite.users["tester"].Id)
	assert.Equal(suite.T(), true, doc.SharedUsers[1].Permissions.Read)
	assert.Equal(suite.T(), true, doc.SharedUsers[1].Permissions.Write)

	userDocs := getDocuments(suite.T(), suite.userHttp, 200)
	adminDocs := getDocuments(suite.T(), suite.adminHttp, 200)
	testerDocs := getDocuments(suite.T(), suite.testerHttp, 200)
	assert.Equal(suite.T(), 6, len(*userDocs), "no documents shared")
	assert.Equal(suite.T(), 2, len(*adminDocs), "no documents shared")
	assert.Equal(suite.T(), 1, len(*testerDocs), "no documents shared")

	request = &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  true,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)

	doc = getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Equal(suite.T(), 1, doc.Shares)
	assert.NotEmpty(suite.T(), doc.SharedUsers)
	assert.Equal(suite.T(), doc.SharedUsers[0].UserId, suite.users["tester"].Id)
	assert.Equal(suite.T(), true, doc.SharedUsers[0].Permissions.Read)
	assert.Equal(suite.T(), true, doc.SharedUsers[0].Permissions.Write)

	userDocs = getDocuments(suite.T(), suite.userHttp, 200)
	adminDocs = getDocuments(suite.T(), suite.adminHttp, 200)
	testerDocs = getDocuments(suite.T(), suite.testerHttp, 200)
	assert.Equal(suite.T(), 6, len(*userDocs), "no documents shared")
	assert.Equal(suite.T(), 1, len(*adminDocs), "no documents shared")
	assert.Equal(suite.T(), 1, len(*testerDocs), "no documents shared")
}

func (suite *ShareDocumentTestSuite) TestReadPermission() {
	doc := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	testersDoc := getDocument(suite.T(), suite.testerHttp, testDocumentMetamorphosis.Id, 404)
	_ = getDocumentHistory(suite.T(), suite.testerHttp, doc.Id, 404)
	assert.Nil(suite.T(), testersDoc)
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)
	testersDocs := getDocuments(suite.T(), suite.testerHttp, 200)
	assert.Equal(suite.T(), 1, len(*testersDocs), "no documents shared")

	testersDoc = getDocument(suite.T(), suite.testerHttp, testDocumentMetamorphosis.Id, 200)
	assert.NotNil(suite.T(), testersDoc)

	history := getDocumentHistory(suite.T(), suite.testerHttp, doc.Id, 200)
	assert.NotNil(suite.T(), history)
}

func (suite *ShareDocumentTestSuite) TestWritePermissions() {
	doc := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)
	doc.Description = "modified"
	updateDocument(suite.T(), suite.testerHttp, doc, 404)
	request = &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  true,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, doc.Id, request, 200)
	updateDocument(suite.T(), suite.testerHttp, doc, 200)

	waitIndexingReady(suite.T(), suite.userHttp, 10)

	updatedDoc := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	assert.Equal(suite.T(), "modified", updatedDoc.Description)
	testersDoc := getDocument(suite.T(), suite.testerHttp, testDocumentMetamorphosis.Id, 200)
	assert.Equal(suite.T(), updatedDoc.Name, testersDoc.Name)
}

func (suite *ShareDocumentTestSuite) TestSearch() {
	filter := map[string]string{
		"q": `measure of the complexit`,
	}
	docs := searchDocuments(suite.T(), suite.testerHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 0, len(docs))

	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["tester"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, testDocumentTransistorCount.Id, request, 200)
	time.Sleep(time.Millisecond * 100)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docs = searchDocuments(suite.T(), suite.testerHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), docs[0].Id, testDocumentTransistorCount.Id)
	assert.Equal(suite.T(), docs[0].Shares, 1)

	request = &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{}}
	updateDocumentSharing(suite.T(), suite.userHttp, testDocumentTransistorCount.Id, request, 200)
	time.Sleep(time.Millisecond * 100)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docs = searchDocuments(suite.T(), suite.testerHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 0, len(docs))
}

func (suite *ShareDocumentTestSuite) TestKeywordOwner() {
	requestProcessingAllUserDocument(suite.T(), suite.userHttp)
	waitNoJobsRunning(suite.T(), suite.db, 10)
	time.Sleep(time.Millisecond * 500)
	waitIndexingReady(suite.T(), suite.userHttp, 10)
	filter := map[string]string{
		"q": `transistor count owner:me`,
	}
	docs := searchDocuments(suite.T(), suite.adminHttp, filter, 1, 10, "name", "ASC", 200)
	assert.Equal(suite.T(), 0, len(docs))
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["admin"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, testDocumentTransistorCount.Id, request, 200)
	waitNoJobsRunning(suite.T(), suite.db, 10)
	time.Sleep(time.Millisecond * 500)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter)

	filter["q"] = `transistor count owner:anyone`
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter, testDocumentTransistorCount)

	filter["q"] = `transistor count owner:others`
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter, testDocumentTransistorCount)

	filter["q"] = `transistor count owner:anyone`
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter, testDocumentTransistorCount)
}

func (suite *ShareDocumentTestSuite) TestKeywordShared() {
	requestProcessingAllUserDocument(suite.T(), suite.userHttp)
	waitNoJobsRunning(suite.T(), suite.db, 10)
	time.Sleep(time.Millisecond * 500)
	waitIndexingReady(suite.T(), suite.userHttp, 10)
	filter := map[string]string{
		"q": `shared:no`,
	}
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter)
	searchDocumentsAndAssertResult(suite.T(), suite.userHttp, filter, testDocumentX86, testDocumentX86Intel, testDocumentJupiterMoons, testDocumentMetamorphosis, testDocumentYear1962, testDocumentTransistorCount)
	request := &aggregates.DocumentUpdateSharingRequest{Users: []aggregates.UserPermissions{
		{
			UserId: suite.users["admin"].Id,
			Permissions: models.Permissions{
				Read:   true,
				Write:  false,
				Delete: false,
			},
		},
	}}
	updateDocumentSharing(suite.T(), suite.userHttp, testDocumentTransistorCount.Id, request, 200)

	waitNoJobsRunning(suite.T(), suite.db, 10)
	time.Sleep(time.Millisecond * 500)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	filter["q"] = `shared:no`
	searchDocumentsAndAssertResult(suite.T(), suite.userHttp, filter, testDocumentX86, testDocumentX86Intel, testDocumentJupiterMoons, testDocumentMetamorphosis, testDocumentYear1962)
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter)

	filter["q"] = `shared:yes`
	searchDocumentsAndAssertResult(suite.T(), suite.userHttp, filter, testDocumentTransistorCount)
	searchDocumentsAndAssertResult(suite.T(), suite.adminHttp, filter, testDocumentTransistorCount)
}

func updateDocumentSharing(t *testing.T, client *httpClient, docId string, input *aggregates.DocumentUpdateSharingRequest, wantHttpStatus int) {
	req := client.Put(fmt.Sprintf("/api/v1/documents/%s/sharing", docId))
	req.Json(t, input).ExpectName(t, "update document sharing", false).e.Status(wantHttpStatus).Done()
}

func requestProcessingAllUserDocument(t *testing.T, client *httpClient) {
	for _, v := range testDocumentIdsUser {
		requestDocumentProcessing(t, client, v, 200)
	}
}
