package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
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

// todo: metadata
