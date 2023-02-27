package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func TestDocumentHistory(t *testing.T) {
	suite.Run(t, new(DocumentHistoryTestSuite))
}

type DocumentHistoryTestSuite struct {
	ApiTestSuite
}

func (suite *DocumentHistoryTestSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T(), suite.db)
	_ = insertTestDocuments(suite.T(), suite.db)
}

func (suite *DocumentHistoryTestSuite) TestInitialHistory() {
	history := getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 1)

	assert.Equal(suite.T(), (*history)[0].Action, "create")
	assert.Equal(suite.T(), (*history)[0].OldValue, "")
	assert.Equal(suite.T(), (*history)[0].NewValue, testDocumentX86.Name)
}

func (suite *DocumentHistoryTestSuite) TestChangeDate() {
	history := getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 1)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	oldDate := doc.Date
	doc.Date = time.Now().AddDate(-1, 0, 0).UnixMilli()
	updateDocument(suite.T(), suite.userHttp, doc, 200)
	newDate := doc.Date
	history = getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 2)

	assert.Equal(suite.T(), (*history)[1].Action, "date")
	assert.Equal(suite.T(), (*history)[1].OldValue, fmt.Sprintf("%d", oldDate/1000))
	assert.Equal(suite.T(), (*history)[1].NewValue, fmt.Sprintf("%d", newDate/1000))
}

func (suite *DocumentHistoryTestSuite) TestRenameDocument() {
	history := getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 1)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	oldName := doc.Name
	doc.Name = "renamed file"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	history = getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 2)

	assert.Equal(suite.T(), (*history)[1].Action, "rename")
	assert.Equal(suite.T(), (*history)[1].OldValue, oldName)
	assert.Equal(suite.T(), (*history)[1].NewValue, doc.Name)
}

func (suite *DocumentHistoryTestSuite) TestChangeDescription() {
	history := getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 1)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	oldDescription := doc.Description
	doc.Description = "new description"
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	history = getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 2)

	assert.Equal(suite.T(), (*history)[1].Action, "description")
	assert.Equal(suite.T(), (*history)[1].OldValue, oldDescription)
	assert.Equal(suite.T(), (*history)[1].NewValue, doc.Description)
}

func (suite *DocumentHistoryTestSuite) TestChangeMetadata() {
	history := getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 1)

	key1 := AddMetadataKey(suite.T(), suite.userHttp, "key", "", 200)
	key2 := AddMetadataKey(suite.T(), suite.userHttp, "key2", "", 200)

	value1 := AddMetadataValue(suite.T(), suite.userHttp, key1.Id, &models.MetadataValue{Value: "value1"}, 200)
	value2 := AddMetadataValue(suite.T(), suite.userHttp, key2.Id, &models.MetadataValue{Value: "value1"}, 200)
	value3 := AddMetadataValue(suite.T(), suite.userHttp, key2.Id, &models.MetadataValue{Value: "value2"}, 200)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	doc.Metadata = []models.Metadata{
		{
			KeyId: key1.Id, ValueId: value1.Id, Key: "", Value: "",
		}}
	updateDocument(suite.T(), suite.userHttp, doc, 200)
	doc = getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	doc.Metadata =
		[]models.Metadata{
			{
				KeyId: key2.Id, ValueId: value2.Id, Key: "", Value: "",
			},
			{
				KeyId: key2.Id, ValueId: value3.Id, Key: "", Value: "",
			},
		}
	updateDocument(suite.T(), suite.userHttp, doc, 200)

	history = getDocumentHistory(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.Len(suite.T(), *history, 5)

	assert.Equal(suite.T(), (*history)[1].Action, "add metadata")
	assert.Equal(suite.T(), (*history)[1].OldValue, "")
	assert.Equal(suite.T(), (*history)[1].NewValue, fmt.Sprintf(`{"key_id":%d,"value_id":%d}`, key1.Id, value1.Id))

	assert.Equal(suite.T(), (*history)[2].Action, "remove metadata")
	assert.Equal(suite.T(), (*history)[2].OldValue, fmt.Sprintf(`{"key_id":%d,"value_id":%d}`, key1.Id, value1.Id))
	assert.Equal(suite.T(), (*history)[2].NewValue, "")

	assert.Equal(suite.T(), (*history)[3].Action, "add metadata")
	assert.Equal(suite.T(), (*history)[3].OldValue, "")
	assert.Equal(suite.T(), (*history)[3].NewValue, fmt.Sprintf(`{"key_id":%d,"value_id":%d}`, key2.Id, value2.Id))

	assert.Equal(suite.T(), (*history)[4].Action, "add metadata")
	assert.Equal(suite.T(), (*history)[4].OldValue, "")
	assert.Equal(suite.T(), (*history)[4].NewValue, fmt.Sprintf(`{"key_id":%d,"value_id":%d}`, key2.Id, value3.Id))
}
