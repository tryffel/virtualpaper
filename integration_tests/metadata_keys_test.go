package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataKeySuite struct {
	ApiTestSuite
	keys map[string]*models.MetadataKey
}

func (suite *MetadataKeySuite) SetupTest() {
	suite.Init()
	suite.keys = make(map[string]*models.MetadataKey)
	suite.keys["author"] = AddMetadataKey(suite.T(), suite.userHttp, "author", "document author", 200)

	suite.keys["category"] = AddMetadataKey(suite.T(), suite.userHttp, "category", "document category", 200)
	suite.keys["test"] = AddMetadataKey(suite.T(), suite.userHttp, "test", "test key", 200)
	suite.keys["admin-test"] = AddMetadataKey(suite.T(), suite.adminHttp, "test", "test key", 200)
	assert.Equal(suite.T(), 4, len(suite.keys))
}

func TestMetadataKeys(t *testing.T) {
	suite.Run(t, new(MetadataKeySuite))
}

func (suite *MetadataKeySuite) TestAddInvalidKeys() {
	AddMetadataKey(suite.T(), suite.userHttp, "", "document author", 400)
	AddMetadataKey(suite.T(), suite.userHttp, "test:key", "document author", 400)
	AddMetadataKey(suite.T(), suite.userHttp, "test;", "document author", 400)
	AddMetadataKey(suite.T(), suite.userHttp, "test\nnewline", "document author", 400)
	// too long key
	AddMetadataKey(suite.T(), suite.userHttp, "LoremIpsumissimplydummytextoftheprintingandtypesetting", "document author", 400)
}

func (suite *MetadataKeySuite) TestUpdateKey() {
	key := suite.keys["author"]
	key.Key = "testing author"

	UpdateMetadataKey(suite.T(), suite.userHttp, 200, key)

	// admin user does not have access to the key
	UpdateMetadataKey(suite.T(), suite.adminHttp, 404, key)

	newKey := GetMetadataKey(suite.T(), suite.userHttp, key.Id, 200)
	assert.Equal(suite.T(), key.Id, newKey.Id, "key id")
	assert.Equal(suite.T(), "testing author", newKey.Key, "key key")
}

func (suite *MetadataKeySuite) TestDeleteKey() {
	key := suite.keys["author"]

	DeleteMetadataKey(suite.T(), suite.adminHttp, 404, key.Id)
	DeleteMetadataKey(suite.T(), suite.userHttp, 200, key.Id)

	DeleteMetadataKey(suite.T(), suite.userHttp, 404, key.Id)
	GetMetadataKey(suite.T(), suite.userHttp, key.Id, 404)
}

func (suite *MetadataKeySuite) TestGetKeys() {
	keys := GetMetadataKeys(suite.T(), suite.userHttp, 200, func(req *httpRequest) *httpRequest {
		return req.Sort("key", "ASC")
	})
	assert.Equal(suite.T(), len(suite.keys)-1, len(*keys), "number of keys match")
	assert.Equal(suite.T(), (*keys)[0].Id, suite.keys["author"].Id, "1st key matches")
	assert.Equal(suite.T(), (*keys)[1].Id, suite.keys["category"].Id, "2nd key matches")
	assert.Equal(suite.T(), (*keys)[2].Id, suite.keys["test"].Id, "3rd key matches")

	keys = GetMetadataKeys(suite.T(), suite.userHttp, 200, func(req *httpRequest) *httpRequest {
		return req.Sort("key", "DESC")
	})

	assert.Equal(suite.T(), len(suite.keys)-1, len(*keys), "number of keys match")
	assert.Equal(suite.T(), (*keys)[0].Id, suite.keys["test"].Id, "1st key matches")
	assert.Equal(suite.T(), (*keys)[1].Id, suite.keys["category"].Id, "2nd key matches")
	assert.Equal(suite.T(), (*keys)[2].Id, suite.keys["author"].Id, "3rd key matches")

	keys = GetMetadataKeys(suite.T(), suite.userHttp, 200, func(req *httpRequest) *httpRequest {
		return req.Sort("created_at", "DESC")
	})

	assert.Equal(suite.T(), len(suite.keys)-1, len(*keys), "number of keys match")
	assert.Equal(suite.T(), (*keys)[0].Id, suite.keys["test"].Id, "1st key matches")
	assert.Equal(suite.T(), (*keys)[1].Id, suite.keys["category"].Id, "2nd key matches")
	assert.Equal(suite.T(), (*keys)[2].Id, suite.keys["author"].Id, "3rd key matches")
}
