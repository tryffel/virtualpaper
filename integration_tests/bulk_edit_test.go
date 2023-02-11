package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type BulkEditTestSuite struct {
	ApiTestSuite
	user1Keys   map[string]*models.MetadataKey
	user1Values map[string]map[string]*models.MetadataValue

	user2Keys   map[string]*models.MetadataKey
	user2Values map[string]map[string]*models.MetadataValue
}

func (suite *BulkEditTestSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T())

	suite.user1Keys, suite.user1Values = initMetadataKeyValues(suite.T(), suite.userHttp)

	suite.user2Keys = make(map[string]*models.MetadataKey)
	suite.user2Keys["author"] = AddMetadataKey(suite.T(), suite.adminHttp, "author", "document author", 200)
	suite.user2Keys["category"] = AddMetadataKey(suite.T(), suite.adminHttp, "category", "document category", 200)
	suite.user2Keys["subject"] = AddMetadataKey(suite.T(), suite.adminHttp, "subject", "document subject", 200)
	assert.Equal(suite.T(), len(suite.user1Keys), 3)
	assert.Equal(suite.T(), len(suite.user2Keys), 3)

	suite.user2Values = make(map[string]map[string]*models.MetadataValue)
	suite.user2Values["author"] = map[string]*models.MetadataValue{}
	suite.user2Values["category"] = map[string]*models.MetadataValue{}
	suite.user2Values["subject"] = map[string]*models.MetadataValue{}

	suite.user2Values["author"]["doyle"] = AddMetadataValue(suite.T(), suite.userHttp, suite.user2Keys["author"].Id, &models.MetadataValue{
		Value:          "doyle",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)
	suite.user2Values["author"]["darwin"] = AddMetadataValue(suite.T(), suite.userHttp, suite.user2Keys["author"].Id, &models.MetadataValue{
		Value:          "darwin",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)

	err := insertTestDocuments(suite.T())
	if err != nil {
		suite.T().Errorf("insert test documents: %v", err)
		suite.T().Fail()
	}
}

func (suite *BulkEditTestSuite) TestAddMetadata() {
	originalX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	originalX86Intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	originalMetamorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)

	assert.Len(suite.T(), originalX86.Metadata, 0)
	assert.Len(suite.T(), originalX86Intel.Metadata, 0)
	assert.Len(suite.T(), originalMetamorphosis.Metadata, 0)

	valueDoyle := suite.user1Values["author"]["doyle"]

	suite.T().Log("bulk edit zero documents, add one key, should fail")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: valueDoyle.Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 400)

	suite.T().Log("bulk edit one document, no edits, should fail")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			originalX86.Id,
		},
		AddMetadata:    api.MetadataUpdateRequest{[]api.MetadataRequest{}},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 304)

	suite.T().Log("bulk edit one document, add one key")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			originalX86.Id,
		},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: valueDoyle.Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 200)

	editedX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	uneditedX86Intel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	uneditedMetamorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)

	assert.Equal(suite.T(), originalX86Intel, uneditedX86Intel)
	assert.Equal(suite.T(), originalMetamorphosis, uneditedMetamorphosis)
	assertDocumentMetadataMatches(suite.T(), editedX86, []*models.MetadataValue{valueDoyle})

	suite.T().Log("bulk edit two documents, add three keys")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			originalX86.Id,
			originalMetamorphosis.Id,
		},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: valueDoyle.Id,
			},
			{
				KeyId:   suite.user1Keys["category"].Id,
				ValueId: suite.user1Values["category"]["paper"].Id,
			},
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: suite.user1Values["author"]["darwin"].Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 200)

	editedX86 = getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	editedMetamorphosis := getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	wantKeys := []*models.MetadataValue{valueDoyle, suite.user1Values["category"]["paper"], suite.user1Values["author"]["darwin"]}
	assertDocumentMetadataMatches(suite.T(), editedX86, wantKeys)
	assertDocumentMetadataMatches(suite.T(), editedMetamorphosis, wantKeys)

	suite.T().Log("bulk edit two documents, add one key, remove two keys")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			originalX86.Id,
			originalMetamorphosis.Id,
		},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["category"].Id,
				ValueId: suite.user1Values["category"]["invoice"].Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["category"].Id,
				ValueId: suite.user1Values["category"]["paper"].Id,
			},
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: suite.user1Values["author"]["darwin"].Id,
			},
		},
		},
	}, 200)

	editedX86 = getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	editedMetamorphosis = getDocument(suite.T(), suite.userHttp, testDocumentMetamorphosis.Id, 200)
	wantKeys = []*models.MetadataValue{valueDoyle, suite.user1Values["category"]["invoice"]}
	assertDocumentMetadataMatches(suite.T(), editedX86, wantKeys)
	assertDocumentMetadataMatches(suite.T(), editedMetamorphosis, wantKeys)

}

func (suite *BulkEditTestSuite) TestPermissionDenied() {
	suite.T().Log("attempt to edit document that's not ours")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			testDocumentTransistorCountAdminUser.Id,
		},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user1Keys["author"].Id,
				ValueId: suite.user1Values["author"]["doyle"].Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 403)

	originalX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)

	suite.T().Log("attempt to add key that's not ours")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			testDocumentX86Intel.Id,
		},
		AddMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user2Keys["author"].Id,
				ValueId: suite.user2Values["author"]["doyle"].Id,
			},
		},
		},
		RemoveMetadata: api.MetadataUpdateRequest{},
	}, 403)

	editedX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), originalX86, editedX86)

	suite.T().Log("attempt to remove key that's not ours")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			testDocumentX86Intel.Id,
		},
		RemoveMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user2Keys["author"].Id,
				ValueId: suite.user2Values["author"]["doyle"].Id,
			},
		},
		},
	}, 403)

	editedX86 = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), originalX86, editedX86)

	suite.T().Log("attempt to add key that does not exist")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			testDocumentX86Intel.Id,
		},
		RemoveMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   10000005,
				ValueId: suite.user2Values["author"]["doyle"].Id,
			},
		},
		},
	}, 403)

	suite.T().Log("attempt to add value that does not exist")
	doBulkEditRequest(suite.T(), suite.userHttp, &api.BulkEditDocumentsRequest{
		Documents: []string{
			testDocumentX86Intel.Id,
		},
		RemoveMetadata: api.MetadataUpdateRequest{[]api.MetadataRequest{
			{
				KeyId:   suite.user2Keys["author"].Id,
				ValueId: 10000005,
			},
		},
		},
	}, 403)

	editedX86 = getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), originalX86, editedX86)
}

func doBulkEditRequest(t *testing.T, client *httpClient, req *api.BulkEditDocumentsRequest, wantHttpStatus int) {
	out := client.Post("/api/v1/documents/bulkEdit").Json(t, req).ExpectName(t, "bulk edit", false)
	out.e.Status(wantHttpStatus).Done()
}

func TestBulkEditDocuments(t *testing.T) {
	suite.Run(t, new(BulkEditTestSuite))
}
