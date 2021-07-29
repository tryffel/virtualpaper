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

package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
)

// RuleStore is storage for user-defined processing rules.
type RuleStore struct {
	db    *sqlx.DB
	cache *cache.Cache

	metadata *MetadataStore
	sq       squirrel.StatementBuilderType
}

func (s *RuleStore) Name() string {
	return "Rules"
}

func (s *RuleStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func newRuleStore(db *sqlx.DB, metadata *MetadataStore) *RuleStore {
	store := &RuleStore{
		db:       db,
		cache:    cache.New(5*time.Minute, time.Minute),
		metadata: metadata,
		sq:       squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
	return store
}

func (s *RuleStore) getRuleCached(id int) *models.Rule {
	cacheRecord, found := s.cache.Get(fmt.Sprintf("rule-%d", id))
	if found {
		rule, ok := cacheRecord.(*models.Rule)
		if ok {
			return rule
		} else {
			s.cache.Delete(fmt.Sprintf("rule-%d", id))
		}
	}
	return nil
}

func (s *RuleStore) setRuleCached(rule *models.Rule) {
	if rule == nil || rule.Id == 0 {
		return
	}
	s.cache.Set(fmt.Sprintf("rule-%d", rule.Id), rule, cache.DefaultExpiration)
}

func (s *RuleStore) GetUserRules(userId int, paging Paging) ([]*models.Rule, error) {
	sql := `
SELECT *
FROM rules
WHERE user_id = $1
OFFSET $2
LIMIT $3;`

	rules := &[]models.Rule{}
	err := s.db.Select(rules, sql, userId, paging.Offset, paging.Limit)
	if err != nil {
		return nil, s.parseError(err, "get user rules")
	}
	ruleArr := make([]*models.Rule, len(*rules))
	for i, _ := range *rules {
		ruleArr[i] = &(*rules)[i]
	}
	err = s.getUserRuleConditionsForRules(userId, ruleArr)
	if err != nil {
		return ruleArr, fmt.Errorf("get conditions: %v", err)
	}
	err = s.getUserRuleActionsForRules(userId, ruleArr)
	if err != nil {
		return ruleArr, fmt.Errorf("get actions: %v", err)
	}
	return ruleArr, nil
}

func (s *RuleStore) GetUserRule(userId, ruleId int) (*models.Rule, error) {
	sql := `
SELECT *
FROM rules
WHERE user_id = $1
AND id = $2;`

	rule := &models.Rule{}
	err := s.db.Get(rule, sql, userId, ruleId)
	if err != nil {
		return rule, s.parseError(err, "get user rule")
	}

	err = s.getUserRuleConditionsForRules(userId, []*models.Rule{rule})
	if err != nil {
		return rule, fmt.Errorf("get conditions: %v", err)
	}
	err = s.getUserRuleActionsForRules(userId, []*models.Rule{rule})
	if err != nil {
		return rule, fmt.Errorf("get actions: %v", err)
	}
	return rule, nil
}

func (s *RuleStore) AddRule(userId int, rule *models.Rule) error {
	err := rule.Validate()
	if err != nil {
		return err
	}

	metadata := make([]models.Metadata, 0, 5)
	for _, v := range rule.Conditions {
		if v.MetadataValue > 0 && v.MetadataKey > 0 {
			m := models.Metadata{
				KeyId:   int(v.MetadataKey),
				ValueId: int(v.MetadataValue),
			}
			metadata = append(metadata, m)
		}
	}
	for _, v := range rule.Actions {
		if v.MetadataValue > 0 && v.MetadataKey > 0 {
			m := models.Metadata{
				KeyId:   int(v.MetadataKey),
				ValueId: int(v.MetadataValue),
			}
			metadata = append(metadata, m)
		}
	}

	if len(metadata) > 0 {
		err := s.metadata.CheckKeyValuesExist(userId, metadata)
		if err != nil {
			return err
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %v", err)
	}

	// Increase remaining rules rule_order by on.
	// Due to unique constraint a temporary value will need to be first set.
	updateSql := `
				UPDATE rules
				SET rule_order = -rule_order
				WHERE user_id = $1 AND rule_order >= $2;
`
	_, err = tx.Exec(updateSql, userId, rule.Order)
	if err != nil {
		tx.Rollback()
		return getDatabaseError(err, s, "increase rule order")
	}

	updateSql = `UPDATE rules
				SET rule_order = -rule_order +1
				WHERE user_id = $1 AND -rule_order >= $2;`

	_, err = tx.Exec(updateSql, userId, rule.Order)
	if err != nil {
		tx.Rollback()
		return getDatabaseError(err, s, "increase rule order")
	}

	// insert rule
	query := s.sq.Insert("rules").
		Columns("user_id", "name", "description", "enabled", "rule_order", "mode").
		Values(userId, rule.Name, rule.Description, rule.Enabled, rule.Order, rule.Mode).
		Suffix("RETURNING \"id\"")
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build insert rule sql: %v", err)
	}

	var id int
	err = tx.Get(&id, sql, args...)
	if err != nil {
		tx.Rollback()
		return getDatabaseError(err, s, "insert rule")
	}

	rule.Id = id
	query = s.sq.Insert("rule_conditions").
		Columns("rule_id", "enabled", "case_insensitive", "inverted_match", "condition_type",
			"is_regex", "value", "metadata_key", "metadata_value")

	for _, v := range rule.Conditions {
		query = query.Values(rule.Id, v.Enabled, v.CaseInsensitive, v.Inverted, v.ConditionType, v.IsRegex, v.Value,
			v.MetadataKey, v.MetadataValue)
	}

	sql, args, err = query.ToSql()
	if err != nil {
		return fmt.Errorf("construct insert conditions sql: %v", err)
	}

	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return getDatabaseError(err, s, "insert rule conditions")
	}

	query = s.sq.Insert("rule_actions").
		Columns("rule_id", "enabled", "on_condition", "action", "value", "metadata_key", "metadata_value")

	for _, v := range rule.Actions {
		query = query.Values(rule.Id, v.Enabled, v.OnCondition, v.Action, v.Value, v.MetadataKey, v.MetadataValue)
	}

	sql, args, err = query.ToSql()
	if err != nil {
		return fmt.Errorf("construct insert actions sql: %v", err)
	}

	_, err = tx.Exec(sql, args...)
	if err != nil {
		tx.Rollback()
		return getDatabaseError(err, s, "insert rule actions")
	}
	return tx.Commit()
}

// GetActiveUresRules returns all active rules (with some limit) for given user.
func (s *RuleStore) GetActiveUserRules(userId int) ([]*models.Rule, error) {

	sql := `
select *
from rules
where user_id = $1
order by rule_order asc
limit $2;`

	rules := &[]models.Rule{}
	err := s.db.Select(rules, sql, userId, config.MaxRulesToProcess)
	if err != nil {
		return nil, s.parseError(err, "get active user rules")
	}
	ruleArr := make([]*models.Rule, len(*rules))

	for i, _ := range *rules {
		ruleArr[i] = &(*rules)[i]
	}

	err = s.getUserRuleConditionsForRules(userId, ruleArr)
	if err != nil {
		return ruleArr, fmt.Errorf("get conditions: %v", err)
	}
	err = s.getUserRuleActionsForRules(userId, ruleArr)
	if err != nil {
		return ruleArr, fmt.Errorf("get actions: %v", err)
	}
	return ruleArr, nil
}

func (s *RuleStore) getUserRuleConditionsForRules(userId int, rules []*models.Rule) error {
	sql := `
SELECT
    rule_conditions.id AS id,
    rule_id,
    rule_conditions.enabled AS enabled,
    case_insensitive,
    inverted_match,
    condition_type,
    is_regex,
    value,
    metadata_key,
    metadata_value
FROM rule_conditions
	LEFT JOIN rules ON rule_conditions.rule_id = rules.id
WHERE rules.user_id = $1
	AND rule_conditions.enabled = TRUE
ORDER BY rule_id, rule_conditions.id ASC;
`

	conditions := &[]models.RuleCondition{}
	err := s.db.Select(conditions, sql, userId)
	if err != nil {
		return s.parseError(err, "get rule conditions")
	}
	mapConditionsToRules(rules, conditions)
	return nil
}

func (s *RuleStore) getUserRuleActionsForRules(userId int, rules []*models.Rule) error {
	actions := &[]models.RuleAction{}
	sql := `
SELECT
    rule_actions.id AS id,
    rule_id,
    rule_actions.enabled AS enabled,
    on_condition,
    rule_actions.value AS value,
    rule_actions.action AS action,
    metadata_key,
    metadata_value
FROM rule_actions
	LEFT JOIN rules ON rule_actions.rule_id = rules.id
WHERE rules.user_id = $1
	AND rule_actions.enabled = TRUE
ORDER BY rule_id, rule_actions.id ASC;
`
	err := s.db.Select(actions, sql, userId)
	if err != nil {
		return s.parseError(err, "get rule conditions")
	}
	mapActionsToRules(rules, actions)
	return nil
}

func mapConditionsToRules(rules []*models.Rule, conditions *[]models.RuleCondition) {
	for i, _ := range rules {
		rule := rules[i]
		rule.Conditions = make([]*models.RuleCondition, 0, 10)
		for conditionI, condition := range *conditions {
			if condition.RuleId == rule.Id {
				rule.Conditions = append(rule.Conditions, &(*conditions)[conditionI])
			}
		}
	}
}

func mapActionsToRules(rules []*models.Rule, actions *[]models.RuleAction) {
	for i, _ := range rules {
		rule := rules[i]
		rule.Actions = make([]*models.RuleAction, 0, 10)
		for actionI, action := range *actions {
			if action.RuleId == rule.Id {
				rule.Actions = append(rule.Actions, &(*actions)[actionI])
			}
		}
	}
}
