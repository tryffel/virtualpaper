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
	"errors"
	"net/http"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

type ProcessingRule struct {
	Id      int    `json:"id" valid:"-"`
	Type    string `json:"type" valid:"-"`
	Filter  string `json:"filter" valid:"-"`
	Comment string `json:"comment" valid:"-"`
	Active  bool   `json:"active" valid:"-"`
	Action  struct {
		MetadataKey   int    `json:"metadata_key_id" valid:"-"`
		MetadataValue int    `json:"metadata_value_id" valid:"-"`
		Tag           int    `json:"tag_id" valid:"-"`
		DateFmt       string `json:"date_fmt" valid:"-"`
		Description   string `json:"description" valid:"-"`
	} `json:"action" valid:"-"`
	CreatedAd int64 `json:"created_at" valid:"-"`
	UpdatedAt int64 `json:"updated_at" valid:"-"`
}

func (p *ProcessingRule) toRule() *models.Rule {
	rule := &models.Rule{
		Filter:  p.Filter,
		Comment: p.Comment,
		Active:  p.Active,
		Action: models.RuleActionConfig{
			Action:          0,
			MetadataKeyId:   p.Action.MetadataKey,
			MetadataValueId: p.Action.MetadataValue,
			Tag:             p.Action.Tag,
			DateFmt:         p.Action.DateFmt,
			Description:     p.Action.Description,
		},
	}

	if p.Type == string(models.RegexRule) {
		rule.Type = models.RegexRule
	} else if p.Type == string(models.ExactRule) {
		rule.Type = models.ExactRule
	}

	return rule
}

func ruleToProcessingRule(rule *models.Rule) *ProcessingRule {
	pr := &ProcessingRule{
		CreatedAd: rule.CreatedAt.Unix() * 1000,
		UpdatedAt: rule.UpdatedAt.Unix() * 1000,
		Id:        rule.Id,
		Type:      string(rule.Type),
		Filter:    rule.Filter,
		Comment:   rule.Comment,
		Action: struct {
			MetadataKey   int    `json:"metadata_key_id" valid:"-"`
			MetadataValue int    `json:"metadata_value_id" valid:"-"`
			Tag           int    `json:"tag_id" valid:"-"`
			DateFmt       string `json:"date_fmt" valid:"-"`
			Description   string `json:"description" valid:"-"`
		}{
			MetadataKey:   rule.Action.MetadataKeyId,
			MetadataValue: rule.Action.MetadataValueId,
			Tag:           rule.Action.Tag,
			DateFmt:       rule.Action.DateFmt,
			Description:   rule.Action.Description,
		},
	}
	return pr
}

func (a *Api) addUserRule(resp http.ResponseWriter, req *http.Request) {
	handler := "api.addUserRule"

	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no user_id in request context"), handler)
		return
	}

	processingRule := &ProcessingRule{}
	err := unMarshalBody(req, processingRule)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	rule := processingRule.toRule()
	err = rule.Validate()
	if err != nil {
		e := storage.ErrInvalid
		e.ErrMsg = err.Error()
		respError(resp, e, handler)
		return
	}

	err = a.db.RuleStore.AddRule(userId, rule)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	processingRule.Id = rule.Id
	respOk(resp, processingRule)
}

func (a *Api) getUserRules(resp http.ResponseWriter, req *http.Request) {
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

	processingRules := make([]*ProcessingRule, len(*rules))
	for i, v := range *rules {
		processingRules[i] = ruleToProcessingRule(&v)
	}

	respResourceList(resp, processingRules, len(processingRules))
}
