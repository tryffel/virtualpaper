package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/services/process"
)

type RuleApiTestSuite struct {
	ApiTestSuite
}

func TestProcessingRulesApi(t *testing.T) {
	suite.Run(t, new(RuleApiTestSuite))
}

func (suite *RuleApiTestSuite) SetupTest() {
	suite.Init()
	clearDbMetadataTables(suite.T(), suite.db)
	clearDbProcessingRuleTables(suite.T(), suite.db)
}

func (suite *RuleApiTestSuite) AddRules() {

}

func (suite *RuleApiTestSuite) TestCreateRule() {
	rule := &api.Rule{
		Name:        "valid rule",
		Description: "",
		Enabled:     false,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
			},
		},
		Actions: []api.RuleAction{
			{
				Action: "description_append",
				Value:  "test",
			},
		},
	}

	addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")

	rule.Name = "match all"
	rule.Mode = "match_all"
	addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_any")

	rule.Name = "invalid_rule"
	rule.Mode = ""
	addRule(suite.T(), suite.userHttp, rule, 400, "invalid rule type")

	rule.Mode = "match_all"
	rule.Name = "no condition_type"
	rule.Conditions[0].ConditionType = ""
	addRule(suite.T(), suite.userHttp, rule, 400, "no condition_type")

	rule.Name = "no regex"
	rule.Conditions[0].ConditionType = "name_contains"
	rule.Conditions[0].IsRegex = false
	rule.Conditions[0].Value = "invalid regex (("
	addRule(suite.T(), suite.userHttp, rule, 200, "valid rule: no regex")

	rule.Name = "invalid regex"
	rule.Conditions[0].ConditionType = "name_contains"
	rule.Conditions[0].IsRegex = true
	rule.Conditions[0].Value = "invalid regex (("
	addRule(suite.T(), suite.userHttp, rule, 400, "invalid regex")

	rule.Name = "valid regex"
	rule.Conditions[0].ConditionType = "name_contains"
	rule.Conditions[0].IsRegex = true
	rule.Conditions[0].Value = "valid regex (ab)"
	addRule(suite.T(), suite.userHttp, rule, 200, "valid regex")

	rule.Name = "date_is"
	rule.Conditions[0].ConditionType = "date_is"
	rule.Conditions[0].IsRegex = true
	rule.Conditions[0].DateFmt = "abcd"
	rule.Conditions[0].Value = "asdf"
	addRule(suite.T(), suite.userHttp, rule, 200, "date_is")

	rule.Name = "invalid date_is"
	rule.Conditions[0].IsRegex = false
	addRule(suite.T(), suite.userHttp, rule, 200, "invalid date_is, no regex")

	rule.Name = "invalid date_is"
	rule.Conditions[0].IsRegex = true
	rule.Conditions[0].Value = "invalid regex (("
	addRule(suite.T(), suite.userHttp, rule, 400, "invalid date_is, invalid regex")

	rule.Name = "valid action"
	rule.Conditions[0].IsRegex = true
	rule.Conditions[0].Value = "invalid regex"
	rule.Actions[0].Action = "name_set"
	addRule(suite.T(), suite.userHttp, rule, 200, "rule ok")

	rule.Name = "invalid action"
	rule.Actions[0].Action = "name"
	// TODO: should fail
	addRule(suite.T(), suite.userHttp, rule, 200, "invalid rule: bad action name")

	rule.Actions[0].Action = "name_set"

	actions := rule.Actions
	conditions := rule.Conditions

	rule.Actions = nil
	// TODO: should return 400
	addRule(suite.T(), suite.userHttp, rule, 500, "invalid rule: no actions")

	rule.Actions = actions
	rule.Conditions = nil
	addRule(suite.T(), suite.userHttp, rule, 500, "invalid rule: no conditions")

	rule.Conditions = conditions
	addRule(suite.T(), suite.userHttp, rule, 200, "ok")
}

func (suite *RuleApiTestSuite) TestUpdateRule() {
	rule := &api.Rule{
		Name:        "valid rule",
		Description: "",
		Enabled:     false,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
				Enabled:       false,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.Equal(suite.T(), gotRule.UpdatedAt, gotRule.CreatedAt, "timestamps match")
	gotRule2 := updateRule(suite.T(), suite.userHttp, gotRule, 200, "")
	assert.NotEqual(suite.T(), gotRule2.UpdatedAt, gotRule2.CreatedAt, "timestamps don't match")

	rule.Description = "changed description"
	rule.Id = gotRule.Id
	gotRule2 = updateRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.Equal(suite.T(), "changed description", gotRule2.Description, "timestamps don't match")

	rule.Description = "enabled rule"
	rule.Enabled = true
	gotRule2 = updateRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.Equal(suite.T(), true, gotRule2.Enabled, "rule enabled")
	assert.Equal(suite.T(), false, gotRule2.Conditions[0].Enabled, "rule condition enabled")
	assert.Equal(suite.T(), false, gotRule2.Actions[0].Enabled, "rule action enabled")

	rule.Conditions[0].Enabled = true
	gotRule2 = updateRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.Equal(suite.T(), true, gotRule2.Conditions[0].Enabled, "rule condition enabled")

	rule.Actions[0].Enabled = true
	gotRule2 = updateRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.Equal(suite.T(), true, gotRule2.Actions[0].Enabled, "rule action enabled")
}

func (suite *RuleApiTestSuite) TestDeleteRule() {
	rule := &api.Rule{
		Name:        "valid rule",
		Description: "",
		Enabled:     false,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
				Enabled:       false,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	addedRule := addRule(suite.T(), suite.userHttp, rule, 200, "")
	assert.NotNil(suite.T(), addedRule, "rule exists")

	deleteRule(suite.T(), suite.adminHttp, rule.Id, 404)

	gotRule := getRule(suite.T(), suite.userHttp, addedRule.Id, 200)
	assert.NotNil(suite.T(), gotRule, "rule exists")

	deleteRule(suite.T(), suite.userHttp, addedRule.Id, 200)
	gotRule = getRule(suite.T(), suite.userHttp, addedRule.Id, 404)
	assert.Nil(suite.T(), gotRule, "rule exists")
}

func (suite *RuleApiTestSuite) TestGetRules() {
	rule := &api.Rule{
		Name:        "rule1",
		Description: "",
		Enabled:     false,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
				Enabled:       false,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	rule1 := addRule(suite.T(), suite.userHttp, rule, 200, "add rule 1")
	rule.Name = "rule2"

	rule2 := addRule(suite.T(), suite.userHttp, rule, 200, "add rule 2")

	rule.Name = "admin rule"
	adminRule := addRule(suite.T(), suite.adminHttp, rule, 200, "add admin rule")

	rules := getRules(suite.T(), suite.adminHttp, 200, func(req *httpRequest) *httpRequest {
		return req.Sort("name", "ASC")
	})

	assert.Equal(suite.T(), 1, len(*rules), "number of rules match")
	assert.Equal(suite.T(), "admin rule", (*rules)[0].Name, "name matches")
	assert.Equal(suite.T(), adminRule.Id, (*rules)[0].Id, "id matches")

	rules = getRules(suite.T(), suite.userHttp, 200, func(req *httpRequest) *httpRequest {
		return req.Sort("name", "ASC")
	})

	assert.Equal(suite.T(), 2, len(*rules), "number of rules match")
	assert.Equal(suite.T(), rule1.Id, (*rules)[0].Id, "id matches")
	assert.Equal(suite.T(), rule2.Id, (*rules)[1].Id, "id matches")
}

func (suite *RuleApiTestSuite) GetRule() {
	rule := &api.Rule{
		Name:        "rule1",
		Description: "",
		Enabled:     false,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
				Enabled:       false,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	rule1 := addRule(suite.T(), suite.userHttp, rule, 200, "add rule 1")
	rule.Name = "rule2"

	rule2 := addRule(suite.T(), suite.userHttp, rule, 200, "add rule 2")

	rule.Name = "admin rule"
	adminRule := addRule(suite.T(), suite.adminHttp, rule, 200, "add admin rule")

	gotRule1 := getRule(suite.T(), suite.adminHttp, rule1.Id, 404)
	assert.Nil(suite.T(), gotRule1, "admin user can't get user's rule by id")

	gotRule1 = getRule(suite.T(), suite.userHttp, rule1.Id, 200)
	assert.NotNil(suite.T(), gotRule1, "user gets rule by id")

	assert.Equal(suite.T(), rule1.Id, gotRule1.Id, "")
	assert.Equal(suite.T(), rule1.Name, gotRule1.Name, "")

	gotRule2 := getRule(suite.T(), suite.userHttp, rule2.Id, 200)
	assert.Equal(suite.T(), rule2.Id, gotRule2.Id, "")
	assert.Equal(suite.T(), rule2.Name, gotRule2.Name, "")

	gotRule := getRule(suite.T(), suite.userHttp, 0, 404)
	assert.Nil(suite.T(), gotRule, "returns 404 on non-existing rule")

	_ = getRule(suite.T(), suite.adminHttp, adminRule.Id, 200)
}

func (suite *RuleApiTestSuite) TestRuleTestingMatch() {
	_ = insertTestDocuments(suite.T(), suite.db)
	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), testDocumentX86Intel.Name, doc.Name)

	suite.T().Log("test rule with 'match_any'")
	rule := &api.Rule{
		Name:        "test rule match any",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "content_contains",
				IsRegex:       true,
				Value:         "[pP]ersonalll",
				Enabled:       true,
			},
			{
				ConditionType: "content_contains",
				IsRegex:       false,
				Value:         "widely used",
				Enabled:       true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	ruleTest := testRule(suite.T(), suite.userHttp, rule.Id, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), true, ruleTest.Match)

	assert.Len(suite.T(), ruleTest.Conditions, 2)
	assert.Len(suite.T(), ruleTest.Actions, 1)

	assert.Len(suite.T(), ruleTest.ConditionOutput, 2)
	assert.Len(suite.T(), ruleTest.ActionOutput, 1)

	assert.Equal(suite.T(), ruleTest.Conditions[0].ConditionId, gotRule.Conditions[0].Id)
	assert.Equal(suite.T(), ruleTest.Conditions[0].ConditionType, gotRule.Conditions[0].ConditionType)
	assert.Equal(suite.T(), ruleTest.Conditions[0].Matched, false)
	assert.Equal(suite.T(), ruleTest.Conditions[0].Skipped, false)

	assert.Equal(suite.T(), ruleTest.Conditions[1].ConditionId, gotRule.Conditions[1].Id)
	assert.Equal(suite.T(), ruleTest.Conditions[1].ConditionType, gotRule.Conditions[1].ConditionType)
	assert.Equal(suite.T(), ruleTest.Conditions[1].Matched, true)
	assert.Equal(suite.T(), ruleTest.Conditions[1].Skipped, false)

	assert.Equal(suite.T(), ruleTest.Actions[0].ActionId, gotRule.Actions[0].Id)
	assert.Equal(suite.T(), ruleTest.Actions[0].ActionType, gotRule.Actions[0].Action)
	assert.Equal(suite.T(), ruleTest.Actions[0].Skipped, true)

	assert.Equal(suite.T(), ruleTest.ConditionOutput[0], []string{"condition didn't match"})
	assert.Equal(suite.T(), ruleTest.ConditionOutput[1], []string{"condition matched",
		"rule mode is set to 'match any', skip rest conditions"})

	assert.Equal(suite.T(), ruleTest.ActionOutput[0], []string{"action is disabled"})
}

func (suite *RuleApiTestSuite) TestRuleTestingNoMatch() {
	_ = insertTestDocuments(suite.T(), suite.db)
	doc := getDocument(suite.T(), suite.userHttp, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), testDocumentX86Intel.Name, doc.Name)

	suite.T().Log("test rule with 'match_all'")
	rule := &api.Rule{
		Name:        "test rule match all",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_all",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "content_contains",
				IsRegex:       true,
				Value:         "[pP]ersonal",
				Enabled:       true,
			},
			{
				ConditionType: "content_contains",
				IsRegex:       false,
				Value:         "not found",
				Enabled:       true,
			},
		},
		Actions: []api.RuleAction{
			{
				Action:  "description_append",
				Value:   "test",
				Enabled: false,
			},
		},
	}

	gotRule := addRule(suite.T(), suite.userHttp, rule, 200, "add rule")
	rule.Id = gotRule.Id

	ruleTest := testRule(suite.T(), suite.userHttp, rule.Id, testDocumentX86Intel.Id, 200)
	assert.Equal(suite.T(), false, ruleTest.Match)

	assert.Len(suite.T(), ruleTest.Conditions, 2)
	assert.Len(suite.T(), ruleTest.Actions, 1)

	assert.Len(suite.T(), ruleTest.ConditionOutput, 2)
	assert.Len(suite.T(), ruleTest.ActionOutput, 0)

	assert.Equal(suite.T(), ruleTest.Conditions[0].ConditionId, gotRule.Conditions[0].Id)
	assert.Equal(suite.T(), ruleTest.Conditions[0].ConditionType, gotRule.Conditions[0].ConditionType)
	assert.Equal(suite.T(), ruleTest.Conditions[0].Matched, true)
	assert.Equal(suite.T(), ruleTest.Conditions[0].Skipped, false)

	assert.Equal(suite.T(), ruleTest.Conditions[1].ConditionId, gotRule.Conditions[1].Id)
	assert.Equal(suite.T(), ruleTest.Conditions[1].ConditionType, gotRule.Conditions[1].ConditionType)
	assert.Equal(suite.T(), ruleTest.Conditions[1].Matched, false)
	assert.Equal(suite.T(), ruleTest.Conditions[1].Skipped, false)

	assert.Equal(suite.T(), ruleTest.ConditionOutput[0], []string{"condition matched"})
	assert.Equal(suite.T(), ruleTest.ConditionOutput[1],
		[]string{
			"condition didn't match",
			"rule mode is set to 'match all', stopping execution",
		})
}

func (suite *RuleApiTestSuite) TestReorderRules() {
	rule := &api.Rule{
		Name:        "rule1",
		Description: "",
		Enabled:     false,
		Order:       0,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
			},
		},
		Actions: []api.RuleAction{
			{
				Action: "description_append",
				Value:  "test",
			},
		},
	}

	rule1 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")
	rule.Name = "rule2"
	rule2 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")

	rule.Name = "rule3"
	rule3 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")

	rule.Name = "rule4"
	rule4 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")

	newOrder := []int{rule1.Id, rule2.Id, rule4.Id, rule3.Id}
	reorderRules(suite.T(), suite.userHttp, newOrder, 200)
	reorderedRules := getRules(suite.T(), suite.userHttp, 200, nil)

	assert.Equal(suite.T(), (*reorderedRules)[0].Id, newOrder[0])
	assert.Equal(suite.T(), (*reorderedRules)[1].Id, newOrder[1])
	assert.Equal(suite.T(), (*reorderedRules)[2].Id, newOrder[2])
	assert.Equal(suite.T(), (*reorderedRules)[3].Id, newOrder[3])
}

func (suite *RuleApiTestSuite) TestReorderRulesErrors() {
	rule := &api.Rule{
		Name:        "rule1",
		Description: "",
		Enabled:     false,
		Order:       0,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "name_contains",
				IsRegex:       true,
				Value:         "valid regex",
			},
		},
		Actions: []api.RuleAction{
			{
				Action: "description_append",
				Value:  "test",
			},
		},
	}

	reorderRules(suite.T(), suite.userHttp, []int{}, 400)
	reorderRules(suite.T(), suite.userHttp, []int{1}, 400)
	reorderRules(suite.T(), suite.userHttp, []int{1, 2}, 404)

	rule1 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")
	rule.Name = "rule2"
	rule2 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")

	rule.Name = "rule3"
	rule3 := addRule(suite.T(), suite.userHttp, rule, 200, "valid rule, match_all")
	originalRules := getRules(suite.T(), suite.userHttp, 200, nil)
	newOrder := []int{rule1.Id, rule3.Id, rule2.Id}
	reorderRules(suite.T(), suite.adminHttp, newOrder, 404)

	// reorder nonexisting rule
	reorderRules(suite.T(), suite.userHttp, append(newOrder, 10105), 404)
	uneditedRules := getRules(suite.T(), suite.userHttp, 200, nil)
	assert.Equal(suite.T(), originalRules, uneditedRules)

}

func addRule(t *testing.T, client *httpClient, rule *api.Rule, expectStatus int, name string) *api.Rule {
	data := &api.Rule{}
	req := client.Post("/api/v1/processing/rules").Json(t, rule).ExpectName(t, name, false)
	if expectStatus == 200 {
		req = req.Json(t, data)
	}
	req.e.Status(expectStatus).Done()
	return data
}

func updateRule(t *testing.T, client *httpClient, rule *api.Rule, expectStatus int, name string) *api.Rule {
	data := &api.Rule{}
	req := client.Put("/api/v1/processing/rules/"+strconv.Itoa(rule.Id)).Json(t, rule).ExpectName(t, name, false)
	if expectStatus == 200 {
		req = req.Json(t, data)
	}
	req.e.Status(expectStatus).Done()
	return data
}

func getRules(t *testing.T, client *httpClient, wantHttpStatus int, editFunc func(request *httpRequest) *httpRequest) *[]api.Rule {
	req := client.Get("/api/v1/processing/rules")
	if editFunc != nil {
		req = editFunc(req)
	}
	dto := &[]api.Rule{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, dto).e.Status(200).Done()
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
	}
	return dto
}

func getRule(t *testing.T, client *httpClient, id int, wantHttpStatus int) *api.Rule {
	req := client.Get("/api/v1/processing/rules/" + strconv.Itoa(id))
	dto := &api.Rule{}
	if wantHttpStatus == 200 {
		req.Expect(t).Json(t, dto).e.Status(200).Done()
		return dto
	} else {
		req.req.Expect(t).Status(wantHttpStatus).Done()
		return nil
	}
}

func deleteRule(t *testing.T, client *httpClient, id int, wantHttpStatus int) {
	req := client.Delete("/api/v1/processing/rules/" + strconv.Itoa(id))
	req.Expect(t).e.Status(wantHttpStatus).Done()
}

func testRule(t *testing.T, client *httpClient, ruleId int, docId string, wantHttpStatus int) *process.RuleTestResult {
	req := client.Put(fmt.Sprintf("/api/v1/processing/rules/%d/test", ruleId)).Json(t, api.RuleTest{DocumentId: docId})
	if wantHttpStatus == 200 {
		result := &process.RuleTestResult{}
		req.Expect(t).Json(t, result).e.Status(200).Done()
		return result
	} else {
		req.Expect(t).e.Status(200).Done()
		return nil
	}
}

func reorderRules(t *testing.T, client *httpClient, rules []int, wantHttpStatus int) {
	data := &api.ReorderRulesRequest{Ids: rules}
	req := client.Put("/api/v1/processing/rules/reorder").Json(t, data).ExpectName(t, "", false)
	if wantHttpStatus == 200 {
		req = req.Json(t, data)
	}
	req.e.Status(wantHttpStatus).Done()
}
