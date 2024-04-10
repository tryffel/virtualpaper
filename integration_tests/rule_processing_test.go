package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
)

type RuleProcessingTestSuite struct {
	ApiTestSuite
}

func TestRulesProcessing(t *testing.T) {
	suite.Run(t, new(RuleProcessingTestSuite))
}

func (suite *RuleProcessingTestSuite) SetupSuite() {
	suite.ApiTestSuite.SetupSuite()
	clearDbMetadataTables(suite.T(), suite.db)
}

func (suite *RuleProcessingTestSuite) TearDownSuite() {
	clearDbMetadataTables(suite.T(), suite.db)
	suite.ApiTestSuite.TearDownSuite()
}

func (suite *RuleProcessingTestSuite) SetupTest() {
	suite.Init()
	clearDbDocumentTables(suite.T(), suite.db)
	clearDbProcessingRuleTables(suite.T(), suite.db)

	insertTestDocuments(suite.T(), suite.db)
}

func (suite *RuleProcessingTestSuite) TestMatchAny() {
	rule := &api.Rule{
		Name:        "test match any",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_any",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				// should match
				ConditionType:   "content_contains",
				IsRegex:         true,
				Value:           "[pP]ersonal",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				// does not match
				ConditionType: "name_is",
				IsRegex:       false,
				Value:         "invalid",
				Enabled:       true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "description_append",
				Value:       " test",
				Enabled:     true,
				OnCondition: true,
			},
			{
				Action:      "description_append",
				Value:       " invalid test",
				Enabled:     false,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	// invalid user
	requestDocumentProcessing(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 404)

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.NotNil(suite.T(), doc)

	assert.Equal(suite.T(), "description - x86 intel test", doc.Description)
	assertDateMatches(suite.T(), testDocumentX86Intel.Date.Unix(), doc.Date/1000)
}

func (suite *RuleProcessingTestSuite) TestMatchAll() {
	rule := &api.Rule{
		Name:        "test match all",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				// should match
				ConditionType:   "content_contains",
				IsRegex:         true,
				Value:           "[pP]ersonal",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				// does not match
				ConditionType: "name_is",
				IsRegex:       false,
				Value:         "invalid",
				Enabled:       true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "description_append",
				Value:       " test",
				Enabled:     true,
				OnCondition: true,
			},
			{
				Action:      "description_append",
				Value:       " invalid test",
				Enabled:     false,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	// invalid user
	requestDocumentProcessing(suite.T(), suite.adminHttp, testDocumentX86Intel.Id, 404)

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.NotNil(suite.T(), doc)

	assert.Equal(suite.T(), testDocumentX86Intel.Description, doc.Description)
	assertDateMatches(suite.T(), testDocumentX86Intel.Date.Unix(), doc.Date/1000)
}

func (suite *RuleProcessingTestSuite) TestExtractDate() {
	rule := &api.Rule{
		Name:        "test extract date",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_any",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "date_is",
				DateFmt:       "02/01/2006",
				IsRegex:       true,
				Value:         "(\\d{1,2}/\\d{1,2}/\\d{4})",
				Enabled:       true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "date_set",
				Value:       "now()",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	// invalid user
	requestDocumentProcessing(suite.T(), suite.adminHttp, testDocumentX86.Id, 404)

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	// date should not match intel document
	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assertDateMatches(suite.T(), testDocumentX86Intel.Date.Unix(), docIntel.Date/1000)

	// date should match x86 document
	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	assert.NotNil(suite.T(), doc)
	assertDateMatches(suite.T(),
		time.Date(2023, 06, 05, 0, 0, 0, 0, time.UTC).Unix(), doc.Date/1000)
}

func (suite *RuleProcessingTestSuite) TestConditionName() {
	rule := &api.Rule{
		Name:        "test condition name",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_is",
				IsRegex:       false,
				Value:         "x86 intel",
				Enabled:       true,
			},
			{
				ConditionType: "name_starts",
				IsRegex:       false,
				Value:         "x86 i",
				Enabled:       true,
			},
			{
				ConditionType: "name_contains",
				IsRegex:       false,
				Value:         "6 int",
				Enabled:       true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         false,
				Value:           "x86 INTEL",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         true,
				Value:           `x\d{2} intel`,
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         false,
				Value:           `cannot match`,
				Enabled:         false,
				CaseInsensitive: true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "name_set",
				Value:       "renamed",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)

	assert.Equal(suite.T(), "renamed", docIntel.Name)
	assert.Equal(suite.T(), docX86.Name, testDocumentX86.Name)
}

func (suite *RuleProcessingTestSuite) TestConditionDescription() {
	rule := &api.Rule{
		Name:        "test condition description",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "description_is",
				IsRegex:       false,
				Value:         "description - x86 intel",
				Enabled:       true,
			},
			{
				ConditionType: "description_starts",
				IsRegex:       false,
				Value:         "descr",
				Enabled:       true,
			},
			{
				ConditionType: "description_contains",
				IsRegex:       false,
				Value:         "ption -",
				Enabled:       true,
			},
			{
				ConditionType:   "description_contains",
				IsRegex:         false,
				Value:           "PTION -",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "description_contains",
				IsRegex:         true,
				Value:           `\w+\s-\sx86`,
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "description_contains",
				IsRegex:         false,
				Value:           `cannot match`,
				Enabled:         false,
				CaseInsensitive: true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "name_set",
				Value:       "renamed",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)

	assert.Equal(suite.T(), "renamed", docIntel.Name)
	assert.Equal(suite.T(), docX86.Name, testDocumentX86.Name)
}

func (suite *RuleProcessingTestSuite) TestConditionTextContent() {
	rule := &api.Rule{
		Name:        "test condition text content",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "content_is",
				IsRegex:       false,
				Value:         testDocumentX86Intel.Content,
				Enabled:       true,
			},
			{
				ConditionType: "content_starts",
				IsRegex:       false,
				Value:         "The x86 arch",
				Enabled:       true,
			},
			{
				ConditionType: "content_contains",
				IsRegex:       false,
				Value:         "ISAs in the world",
				Enabled:       true,
			},
			{
				ConditionType:   "content_contains",
				IsRegex:         false,
				Value:           "isas in the world -",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "content_contains",
				IsRegex:         true,
				Value:           `\d{4} for use in`,
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "content_contains",
				IsRegex:         false,
				Value:           `cannot match`,
				Enabled:         false,
				CaseInsensitive: true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "name_set",
				Value:       "renamed",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)

	assert.Equal(suite.T(), "renamed", docIntel.Name)
	assert.Equal(suite.T(), docX86.Name, testDocumentX86.Name)
}

func (suite *RuleProcessingTestSuite) TestConditionMetadata() {
	key1 := AddMetadataKey(suite.T(), suite.userHttp, "key1", "", 200)
	key2 := AddMetadataKey(suite.T(), suite.userHttp, "key2", "", 200)

	value1 := AddMetadataValue(suite.T(), suite.userHttp, key1.Id, &models.MetadataValue{Value: "value1"}, 200)
	value2 := AddMetadataValue(suite.T(), suite.userHttp, key1.Id, &models.MetadataValue{Value: "value2"}, 200)
	value3 := AddMetadataValue(suite.T(), suite.userHttp, key2.Id, &models.MetadataValue{Value: "value3"}, 200)

	rule := &api.Rule{
		Name:        "test condition metadata",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "metadata_has_key",
				Metadata:      models.Metadata{KeyId: key1.Id, ValueId: 0},
				Enabled:       true,
			},
			{
				ConditionType: "metadata_has_key_value",
				Metadata:      models.Metadata{KeyId: key1.Id, ValueId: value1.Id},
				Enabled:       true,
			},
			{
				ConditionType: "metadata_has_key_value",
				Metadata:      models.Metadata{KeyId: key2.Id, ValueId: value3.Id},
				Enabled:       true,
				Inverted:      true,
			},
			{
				ConditionType: "metadata_count",
				Value:         "2",
				Enabled:       true,
			},
			{
				ConditionType: "metadata_count_more_than",
				Value:         "5",
				Inverted:      true,
				Enabled:       true,
			},
			{
				ConditionType: "metadata_count_more_than",
				Value:         "0",
				Inverted:      false,
				Enabled:       true,
			},
			{
				ConditionType: "metadata_count_less_than",
				Value:         "3",
				Enabled:       true,
			},
			{
				ConditionType: "metadata_count_less_than",
				Value:         "2",
				Enabled:       true,
				Inverted:      true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "name_set",
				Value:       "renamed",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	doc.Metadata =
		[]models.Metadata{
			{KeyId: key1.Id, ValueId: value1.Id},
			{KeyId: key1.Id, ValueId: value2.Id},
		}
	updateDocument(suite.T(), suite.userHttp, doc, 200)
	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)

	assert.Equal(suite.T(), "renamed", docIntel.Name)
	assert.Equal(suite.T(), docX86.Name, testDocumentX86.Name)
}

func (suite *RuleProcessingTestSuite) TestDocumentCreate() {
	rule := &api.Rule{
		Name:        "test condition name",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Triggers:    models.RuleTriggerArray{"document-create", "document-update"},
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_is",
				IsRegex:       false,
				Value:         "x86 intel",
				Enabled:       true,
			},
			{
				ConditionType: "name_starts",
				IsRegex:       false,
				Value:         "x86 i",
				Enabled:       true,
			},
			{
				ConditionType: "name_contains",
				IsRegex:       false,
				Value:         "6 int",
				Enabled:       true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         false,
				Value:           "x86 INTEL",
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         true,
				Value:           `x\d{2} intel`,
				Enabled:         true,
				CaseInsensitive: true,
			},
			{
				ConditionType:   "name_is",
				IsRegex:         false,
				Value:           `cannot match`,
				Enabled:         false,
				CaseInsensitive: true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:      "name_set",
				Value:       "renamed",
				Enabled:     true,
				OnCondition: true,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86.Id, 200)
	requestDocumentProcessing(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	time.Sleep(time.Second)
	waitIndexingReady(suite.T(), suite.userHttp, 10)

	docIntel := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	docX86 := getDocument(suite.T(), suite.userHttp, testDocumentX86.Id, 200)

	assert.Equal(suite.T(), "renamed", docIntel.Name)
	assert.Equal(suite.T(), docX86.Name, testDocumentX86.Name)
}

func requestDocumentProcessing(t *testing.T, client *httpClient, docId string, expectStatus int) {
	client.Post(fmt.Sprintf("/api/v1/documents/%s/process", docId)).Expect(t).e.Status(expectStatus).Done()
}
