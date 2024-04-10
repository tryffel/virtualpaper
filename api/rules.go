/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2020  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type Rule struct {
	Id          int                     `json:"id" valid:"-"`
	Name        string                  `json:"name" valid:"-"`
	Description string                  `json:"description" valid:"-"`
	Enabled     bool                    `json:"enabled" valid:"-"`
	Order       int                     `json:"order" valid:"-"`
	Mode        string                  `json:"mode" valid:"-"`
	CreatedAt   int64                   `json:"created_at" valid:"-"`
	UpdatedAt   int64                   `json:"updated_at" valid:"-"`
	Triggers    models.RuleTriggerArray `json:"triggers" valid:"rule_trigger_type"`

	Conditions []RuleCondition `json:"conditions" valid:"-"`
	Actions    []RuleAction    `json:"actions" valid:"-"`
}

type RuleCondition struct {
	Id              int             `json:"id" valid:"-"`
	RuleId          int             `json:"rule_id" valid:"-"`
	Enabled         bool            `json:"enabled" valid:"-"`
	CaseInsensitive bool            `json:"case_insensitive" valid:"-"`
	Inverted        bool            `json:"inverted_match" valid:"-"`
	ConditionType   string          `json:"condition_type" valid:"-"`
	IsRegex         bool            `json:"is_regex" valid:"-"`
	Value           string          `json:"value" valid:"-"`
	DateFmt         string          `json:"date_fmt" valid:"-"`
	Metadata        models.Metadata `json:"metadata" valid:"-"`
}

type RuleAction struct {
	Id          int             `json:"id" valid:"-"`
	RuleId      int             `json:"rule_id" valid:"-"`
	Enabled     bool            `json:"enabled" valid:"-"`
	OnCondition bool            `json:"on_condition" valid:"-"`
	Action      string          `json:"action" valid:"-"`
	Value       string          `json:"value" valid:"-"`
	Metadata    models.Metadata `json:"metadata" valid:"-"`
}

type RuleTest struct {
	DocumentId string `json:"document_id" valid:"required"`
}

func (r *RuleAction) ToAction() *models.RuleAction {
	return &models.RuleAction{
		Enabled:       r.Enabled,
		OnCondition:   r.OnCondition,
		Action:        models.RuleActionType(r.Action),
		Value:         r.Value,
		MetadataKey:   models.IntId(r.Metadata.KeyId),
		MetadataValue: models.IntId(r.Metadata.ValueId),
	}
}

func actionToResp(action *models.RuleAction) RuleAction {
	return RuleAction{
		Id:          action.Id,
		RuleId:      action.RuleId,
		Enabled:     action.Enabled,
		OnCondition: action.OnCondition,
		Action:      action.Action.String(),
		Value:       action.Value,
		Metadata: models.Metadata{
			KeyId:   int(action.MetadataKey),
			Key:     action.MetadataKeyName.String(),
			ValueId: int(action.MetadataValue),
			Value:   action.MetadataValueName.String(),
		},
	}
}

func (r *RuleCondition) ToCondition() *models.RuleCondition {
	return &models.RuleCondition{
		Enabled:         r.Enabled,
		CaseInsensitive: r.CaseInsensitive,
		Inverted:        r.Inverted,
		ConditionType:   models.RuleConditionType(r.ConditionType),
		IsRegex:         r.IsRegex,
		Value:           r.Value,
		DateFmt:         r.DateFmt,
		MetadataKey:     models.IntId(r.Metadata.KeyId),
		MetadataValue:   models.IntId(r.Metadata.ValueId),
	}
}

func conditionToResp(cond *models.RuleCondition) RuleCondition {
	return RuleCondition{
		Id:              cond.Id,
		RuleId:          cond.RuleId,
		Enabled:         cond.Enabled,
		CaseInsensitive: cond.CaseInsensitive,
		Inverted:        cond.Inverted,
		ConditionType:   cond.ConditionType.String(),
		IsRegex:         cond.IsRegex,
		Value:           cond.Value,
		DateFmt:         cond.DateFmt,

		Metadata: models.Metadata{
			KeyId:   int(cond.MetadataKey),
			Key:     cond.MetadataKeyName.String(),
			ValueId: int(cond.MetadataValue),
			Value:   cond.MetadataValueName.String(),
		},
	}
}

func ruleToResp(rule *models.Rule) *Rule {
	resp := &Rule{
		Id:          rule.Id,
		Name:        rule.Name,
		Description: rule.Description,
		Enabled:     rule.Enabled,
		Order:       rule.Order,
		Mode:        rule.Mode.String(),
		Triggers:    rule.Triggers,
		CreatedAt:   rule.CreatedAt.Unix() * 1000,
		UpdatedAt:   rule.UpdatedAt.Unix() * 1000,
	}

	resp.Conditions = make([]RuleCondition, len(rule.Conditions))
	resp.Actions = make([]RuleAction, len(rule.Actions))

	for i, v := range rule.Conditions {
		resp.Conditions[i] = conditionToResp(v)
	}
	for i, v := range rule.Actions {
		resp.Actions[i] = actionToResp(v)
	}
	return resp
}

func (r *Rule) ToRule() (*models.Rule, error) {
	mode := models.RuleMatchAll
	err := mode.FromString(r.Mode)
	if err != nil {
		return nil, err
	}

	rule := &models.Rule{
		Name:        r.Name,
		Description: r.Description,
		Enabled:     r.Enabled,
		Order:       r.Order,
		Mode:        mode,
		Conditions:  make([]*models.RuleCondition, len(r.Conditions)),
		Actions:     make([]*models.RuleAction, len(r.Actions)),
		Triggers:    r.Triggers,
	}

	for i, v := range r.Conditions {
		rule.Conditions[i] = v.ToCondition()
	}

	for i, v := range r.Actions {
		rule.Actions[i] = v.ToAction()
	}
	return rule, nil
}

func (a *Api) addUserRule(c echo.Context) error {
	// swagger:route POST /api/v1/processing/rules Processing AddRule
	// Add processing rule
	// responses:
	//   200: ProcessingRuleResponse
	//   304: RespNotModified
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)
	processingRule := &Rule{}
	err := unMarshalBody(c.Request(), processingRule)
	if err != nil {
		return err
	}

	opOk := false
	ruleId := 0
	defer func() {
		logCrudRule(ctx.UserId, "create", &opOk, "rule: %d", ruleId)
	}()

	rule, err := processingRule.ToRule()
	if err != nil {
		return err
	}

	rule.UserId = ctx.UserId

	newRule, err := a.ruleService.Create(getContext(c), rule)
	if err != nil {
		return err
	}
	ruleId = newRule.Id
	opOk = true
	return c.JSON(http.StatusOK, ruleToResp(newRule))
}

func (a *Api) getUserRules(c echo.Context) error {
	// swagger:route GET /api/v1/processing/rules Processing GetRules
	// Get processing rules
	// responses:
	//   200: ProcessingRuleResponse

	ctx := c.(UserContext)

	paging := getPagination(c)

	query, enabledStr, err := getRuleFilter(c.Request())
	if err != nil {
		return err
	}

	rules, total, err := a.ruleService.GetRules(getContext(c), ctx.UserId, paging.toPagination(), strings.ToLower(query), strings.ToLower(enabledStr))
	if err != nil {
		return err
	}

	processingRules := make([]*Rule, len(rules))
	for i, v := range rules {
		processingRules[i] = ruleToResp(v)
	}
	return resourceList(c, processingRules, total)
}

func (a *Api) getUserRule(c echo.Context) error {
	// swagger:route GET /api/v1/processing/rules/{id} Processing GetRule
	// Get processing rule by id
	// responses:
	//   200: ProcessingRuleResponse

	ctx := c.(UserContext)
	id, err := bindPathIdInt(c)
	rule, err := a.ruleService.Get(getContext(c), ctx.UserId, id)
	if err != nil {
		return err
	}

	r := ruleToResp(rule)
	return resourceList(c, r, 1)
}

func (a *Api) updateUserRule(c echo.Context) error {
	// swagger:route PUT /api/v1/processing/rules/{id} Processing UpdateRule
	// UpdateJob rule contents
	// responses:
	//   200:
	ctx := c.(UserContext)
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	processingRule := &Rule{}
	err = unMarshalBody(c.Request(), processingRule)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudRule(ctx.UserId, "update", &opOk, "rule: %d", processingRule.Id)

	rule, err := processingRule.ToRule()
	if err != nil {
		return err
	}
	rule.Id = id
	rule.UserId = ctx.UserId
	err = a.ruleService.Update(getContext(c), rule)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(http.StatusOK, ruleToResp(rule))
}

func (a *Api) deleteUserRule(c echo.Context) error {
	// swagger:route DELETE /api/v1/processing/rules/{id} Processing DeleteRule
	// Delete rule
	// responses:
	//   200:
	ctx := c.(UserContext)
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudRule(ctx.UserId, "update", &opOk, "rule: %d", id)
	err = a.ruleService.Delete(getContext(c), id)
	if err != nil {
		return err
	}
	opOk = true
	return c.String(http.StatusOK, "")
}

func (a *Api) testRule(c echo.Context) error {
	// swagger:route PUT /api/v1/processing/rules/{id}/test Processing TestRule
	// Test rule execution
	// responses:
	//   200: process.RuleTestResult
	//   403:

	ctx := c.(UserContext)
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	processingRule := &RuleTest{}
	err = unMarshalBody(c.Request(), processingRule)
	if err != nil {
		return err

	}

	opOk := false
	matched := false

	defer func() {
		logCrudRule(ctx.UserId, "test", &opOk, "rule: %d, document: %s, matched: %v", id, processingRule.DocumentId, matched)
	}()

	status, err := a.ruleService.TestRule(getContext(c), ctx.UserId, id, processingRule.DocumentId)
	if err != nil {
		return err
	}
	opOk = true
	matched = status.Match
	return c.JSON(http.StatusOK, status)
}

type ReorderRulesRequest struct {
	Ids []int `json:"ids" valid:"-"`
}

func (a *Api) reorderRules(c echo.Context) error {
	ctx := c.(UserContext)
	processingRule := &ReorderRulesRequest{}
	err := unMarshalBody(c.Request(), processingRule)
	if err != nil {
		return err
	}

	if len(processingRule.Ids) < 2 {
		e := errors.ErrInvalid
		e.ErrMsg = "must have at least two rules"
		return e
	}
	if len(processingRule.Ids) > config.MaxRows {
		e := errors.ErrInvalid
		e.ErrMsg = fmt.Sprintf("must have max %d rules", config.MaxRows)
	}

	opOk := false
	defer func() {
		logCrudRule(ctx.UserId, "reorder", &opOk, "")
	}()

	err = a.ruleService.Reorder(getContext(c), ctx.UserId, processingRule.Ids)
	if err != nil {
		return err
	}
	opOk = true
	out := map[string]interface{}{"id": "rules"}
	return c.JSON(200, out)
}
