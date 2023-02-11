package integrationtest

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataValueSuite struct {
	ApiTestSuite
	keys   map[string]*models.MetadataKey
	values map[string]*models.MetadataValue
}

func TestMetadataValue(t *testing.T) {
	suite.Run(t, new(MetadataValueSuite))
}

func (suite *MetadataValueSuite) SetupTest() {
	suite.Init()
	suite.keys = make(map[string]*models.MetadataKey)
	suite.keys["author"] = AddMetadataKey(suite.T(), suite.userHttp, "author", "document author", 200)
	suite.keys["category"] = AddMetadataKey(suite.T(), suite.userHttp, "category", "document category", 200)
	suite.keys["test"] = AddMetadataKey(suite.T(), suite.userHttp, "test", "test key", 200)
	assert.Equal(suite.T(), len(suite.keys), 3)

	AddMetadataValue(suite.T(), suite.userHttp, suite.keys["author"].Id, &models.MetadataValue{
		Value:          "doyle",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)
	AddMetadataValue(suite.T(), suite.userHttp, suite.keys["author"].Id, &models.MetadataValue{
		Value:          "darwin",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)
}

func (suite *MetadataValueSuite) TestInvalidValues() {

	AddMetadataValue(suite.T(), suite.userHttp, suite.keys["author"].Id, &models.MetadataValue{
		Value:          "test:",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 400)

	AddMetadataValue(suite.T(), suite.userHttp, suite.keys["author"].Id, &models.MetadataValue{
		Value:          "test;",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 400)

	AddMetadataValue(suite.T(), suite.userHttp, suite.keys["author"].Id, &models.MetadataValue{
		Value:          "test:second",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 400)

}

func initMetadataKeyValues(t *testing.T, client *httpClient) (map[string]*models.MetadataKey, map[string]map[string]*models.MetadataValue) {
	keys := make(map[string]*models.MetadataKey)
	keys["author"] = AddMetadataKey(t, client, "author", "document author", 200)
	keys["category"] = AddMetadataKey(t, client, "category", "document category", 200)
	keys["test"] = AddMetadataKey(t, client, "test", "test key", 200)

	values := make(map[string]map[string]*models.MetadataValue)
	values["author"] = map[string]*models.MetadataValue{}
	values["category"] = map[string]*models.MetadataValue{}
	values["test"] = map[string]*models.MetadataValue{}

	doyle := AddMetadataValue(t, client, keys["author"].Id, &models.MetadataValue{
		Value:          "doyle",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)
	darwin := AddMetadataValue(t, client, keys["author"].Id, &models.MetadataValue{
		Value:          "darwin",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)

	paper := AddMetadataValue(t, client, keys["category"].Id, &models.MetadataValue{
		Value:          "paper",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)
	invoice := AddMetadataValue(t, client, keys["category"].Id, &models.MetadataValue{
		Value:          "invoice",
		MatchDocuments: false,
		MatchType:      "",
		MatchFilter:    "",
	}, 200)

	values["author"]["doyle"] = doyle
	values["author"]["darwin"] = darwin
	values["category"]["paper"] = paper
	values["category"]["invoice"] = invoice
	return keys, values
}
