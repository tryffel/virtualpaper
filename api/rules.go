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
	"net/http"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type Rule struct {
	Id          int    `json:"id" valid:"-"`
	Name        string `json:"name" valid:"-"`
	Description string `json:"description" valid:"-"`
	Enabled     bool   `json:"enabled" valid:"-"`
	Order       int    `json:"order" valid:"-"`
	Mode        string `json:"mode" valid:"-"`
	CreatedAt   int64  `json:"created_at" valid:"-"`
	UpdatedAt   int64  `json:"updated_at" valid:"-"`

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
	}

	for i, v := range r.Conditions {
		rule.Conditions[i] = v.ToCondition()
	}

	for i, v := range r.Actions {
		rule.Actions[i] = v.ToAction()
	}
	return rule, nil
}

func (a *Api) addUserRule(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/processing/rules Processing AddRule
	// Add processing rule
	// responses:
	//   200: ProcessingRuleResponse
	//   304: RespNotModified
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	handler := "api.addUserRule"

	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	processingRule := &Rule{}
	err := unMarshalBody(req, processingRule)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	rule, err := processingRule.ToRule()
	if err != nil {
		respError(resp, err, handler)
		return
	}

	err = a.db.RuleStore.AddRule(userId, rule)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	rule, err = a.db.RuleStore.GetUserRule(userId, rule.Id)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	respOk(resp, ruleToResp(rule))
}

func (a *Api) getUserRules(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/processing/rules Processing GetRules
	// Get processing rules
	// responses:
	//   200: ProcessingRuleResponse
	handler := "api.getUserRules"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	paging, err := getPaging(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	rules, err := a.db.RuleStore.GetUserRules(userId, paging)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	processingRules := make([]*Rule, len(rules))
	for i, v := range rules {
		processingRules[i] = ruleToResp(v)
	}
	respResourceList(resp, processingRules, len(processingRules))
}

func (a *Api) getUserRule(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/processing/rules/{id} Processing GetRule
	// Get processing rule by id
	// responses:
	//   200: ProcessingRuleResponse
	handler := "api.getUserRule"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	id, err := getParamIntId(req)
	if err != nil {
		respBadRequest(resp, "no id specified", nil)
		return
	}

	rule, err := a.db.RuleStore.GetUserRule(userId, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	r := ruleToResp(rule)
	respResourceList(resp, r, 1)
}

func (a *Api) updateUserRule(resp http.ResponseWriter, req *http.Request) {
	// swagger:route PUT /api/v1/processing/rules/{id} Processing UpdateRule
	// Update rule contents
	// responses:
	//   200:
	handler := "api.updateUserRule"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	id, err := getParamIntId(req)
	if err != nil {
		respBadRequest(resp, "no id specified", nil)
		return
	}

	processingRule := &Rule{}
	err = unMarshalBody(req, processingRule)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	rule, err := processingRule.ToRule()
	if err != nil {
		respError(resp, err, handler)
		return
	}
	rule.Id = id
	rule.UserId = userId
	err = a.db.RuleStore.UpdateRule(userId, rule)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	respOk(resp, ruleToResp(rule))
}

func (a *Api) deleteUserRule(resp http.ResponseWriter, req *http.Request) {
	// swagger:route DELETE /api/v1/processing/rules/{id} Processing DeleteRule
	// Delete rule
	// responses:
	//   200:
	handler := "api.deleteUserRule"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	id, err := getParamIntId(req)
	if err != nil {
		respBadRequest(resp, "no id specified", nil)
		return
	}

	err = a.db.RuleStore.DeleteRule(userId, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	respOk(resp, nil)
}
