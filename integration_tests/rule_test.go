package integrationtest

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"strconv"
	"testing"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/process"
)

type RuleTestSuite struct {
	ApiTestSuite
}

func TestProcessingRules(t *testing.T) {
	suite.Run(t, new(RuleTestSuite))
}

func (suite *RuleTestSuite) SetupTest() {
	suite.Init()
	clearDbMetadataTables(suite.T())
	clearDbProcessingRuleTables(suite.T())
}

func (suite *RuleTestSuite) AddRules() {

}

func (suite *RuleTestSuite) TestCreateRule() {
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

func (suite *RuleTestSuite) TestUpdateRule() {
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

func (suite *RuleTestSuite) TestDeleteRule() {
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

func (suite *RuleTestSuite) TestGetRules() {
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
	assert.Equal(suite.T(), rule1.Id, (*rules)[1].Id, "id matches")
	assert.Equal(suite.T(), rule2.Id, (*rules)[0].Id, "id matches")
}

func (suite *RuleTestSuite) GetRule() {
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

func (suite *RuleTestSuite) TestRuleTesting() {
	docId := uploadDocument(suite.T(), suite.userClient, "text-1.txt", "Lorem ipsum", 20)
	doc := getDocument(suite.T(), suite.userHttp, docId, 200)

	assert.Equal(suite.T(), "file", doc.Name)

	rule := &api.Rule{
		Name:        "rule1",
		Description: "",
		Enabled:     true,
		Order:       10,
		Mode:        "match_any",
		Conditions: []api.RuleCondition{
			{
				ConditionType: "content_contains",
				IsRegex:       true,
				Value:         "[lL]orem ipsum",
				Enabled:       true,
			},
			{
				ConditionType: "content_contains",
				IsRegex:       false,
				Value:         "Lorem ipsum",
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

	ruleTest := testRule(suite.T(), suite.userHttp, rule.Id, docId, 200)
	assert.Equal(suite.T(), true, ruleTest.Match)

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
