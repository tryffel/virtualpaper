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
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

// RuleStore is storage for user-defined processing rules.
type RuleStore struct {
	*resource
	cache *cache.Cache

	metadata *MetadataStore
	sq       squirrel.StatementBuilderType
}

func (s *RuleStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func newRuleStore(db *sqlx.DB, metadata *MetadataStore) *RuleStore {
	store := &RuleStore{
		resource: &resource{name: "Rule", db: db},
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
ORDER BY rule_order ASC
OFFSET $2
LIMIT $3
;`

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
	err := s.validateRule(userId, rule)
	if err != nil {
		return err
	}

	tx, err := s.beginTx()
	defer tx.Close()
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
	_, err = tx.tx.Exec(updateSql, userId, rule.Order)
	if err != nil {
		return getDatabaseError(err, s, "increase rule order")
	}

	updateSql = `UPDATE rules
				SET rule_order = -rule_order +1
				WHERE user_id = $1 AND -rule_order >= $2;`

	_, err = tx.tx.Exec(updateSql, userId, rule.Order)
	if err != nil {
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
	err = tx.tx.Get(&id, sql, args...)
	if err != nil {
		return getDatabaseError(err, s, "insert rule")
	}
	rule.Id = id
	err = s.addActionsToRule(tx, rule.Id, rule.Actions)
	if err != nil {
		return fmt.Errorf("add actions: %v", err)
	}

	err = s.addConditionsToRule(tx, rule.Id, rule.Conditions)
	if err != nil {
		return fmt.Errorf("add conditions: %v", err)
	}
	tx.ok = true
	return nil
}

// GetActiveUresRules returns all enabled rules (with some limit) for given user.
func (s *RuleStore) GetActiveUserRules(userId int) ([]*models.Rule, error) {

	sql := `
SELECT *
FROM rules
WHERE user_id = $1
AND enabled=TRUE
ORDER BY rule_order ASC
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
	rule_conditions.value as value,
    metadata_key,
    metadata_value,
    mk.key as metadata_key_name,
    mv.value as metadata_value_name,
	date_fmt
FROM rule_conditions
	LEFT JOIN rules ON rule_conditions.rule_id = rules.id
	LEFT join metadata_keys mk on rule_conditions.metadata_key = mk.id
	LEFT JOIN metadata_values mv on rule_conditions.metadata_value = mv.id
WHERE rules.user_id = $1
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
    metadata_value,
	mk.key as metadata_key_name,
    mv.value as metadata_value_name
FROM rule_actions
	LEFT JOIN rules ON rule_actions.rule_id = rules.id
	LEFT join metadata_keys mk on rule_actions.metadata_key = mk.id
    LEFT JOIN metadata_values mv on rule_actions.metadata_value = mv.id
WHERE rules.user_id = $1
ORDER BY rule_id, rule_actions.id ASC;
`
	err := s.db.Select(actions, sql, userId)
	if err != nil {
		return s.parseError(err, "get rule conditions")
	}
	mapActionsToRules(rules, actions)
	return nil
}

func (s *RuleStore) UserOwnsRule(userId, ruleId int) (bool, error) {
	sql := `
	SELECT id 
	FROM rules
	WHERE user_id = $1
	AND id = $2;
	`

	var id int
	err := s.db.Get(&id, sql, userId, ruleId)
	if err != nil {
		return false, getDatabaseError(err, s, "check user owns rule")
	}
	return id == ruleId, nil
}

// UpdateRule updates rule.
func (s *RuleStore) UpdateRule(userId int, rule *models.Rule) error {
	owns, err := s.UserOwnsRule(userId, rule.Id)
	if err != nil {
		return err
	}
	if !owns {
		err := errors.ErrRecordNotFound
		err.ErrMsg = "rule not found"
		return err
	}

	err = s.validateRule(userId, rule)
	if err != nil {
		return err
	}

	tx, err := s.beginTx()
	if err != nil {
		return err
	}
	defer tx.Close()

	rule.Update()
	query := s.sq.Update("rules").SetMap(map[string]interface{}{
		"name":        rule.Name,
		"description": rule.Description,
		"enabled":     rule.Enabled,
		"rule_order":  rule.Order,
		"mode":        rule.Mode,
		"updated_at":  rule.UpdatedAt,
	}).Where(squirrel.Eq{"user_id": userId, "id": rule.Id})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %v", err)
	}

	_, err = tx.tx.Exec(sql, args...)
	if err != nil {
		return getDatabaseError(err, s, "update")
	}

	sql = `DELETE FROM rule_actions WHERE rule_id = $1`
	_, err = tx.tx.Exec(sql, rule.Id)
	if err != nil {
		return fmt.Errorf("delete old actions: %v", err)
	}

	sql = `DELETE FROM rule_conditions WHERE rule_id = $1`
	_, err = tx.tx.Exec(sql, rule.Id)
	if err != nil {
		return fmt.Errorf("delete old conditions: %v", err)
	}

	err = s.addActionsToRule(tx, rule.Id, rule.Actions)
	if err != nil {
		return fmt.Errorf("add actions: %v", err)
	}

	err = s.addConditionsToRule(tx, rule.Id, rule.Conditions)
	if err != nil {
		return fmt.Errorf("add conditions: %v", err)
	}

	//TODO: handle changing rule_order

	tx.ok = true
	return nil
}

func (s *RuleStore) DeleteRule(userId, ruleId int) error {
	sql := `
	DELETE FROM rules 
	WHERE user_id = $1
	AND id = $2
	`

	res, err := s.db.Exec(sql, userId, ruleId)
	if err != nil {
		return getDatabaseError(err, s, "delete")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %v", err)
	}
	if affected == 0 {
		err := errors.ErrRecordNotFound
		err.ErrMsg = "rule not found"
		return err
	}
	return nil
}

func (s *RuleStore) validateRule(userId int, rule *models.Rule) error {
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
	return nil
}

func (s *RuleStore) addActionsToRule(tx *tx, ruleId int, actions []*models.RuleAction) error {
	query := s.sq.Insert("rule_actions").
		Columns("rule_id", "enabled", "on_condition", "action", "value", "metadata_key", "metadata_value")

	for _, v := range actions {
		query = query.Values(ruleId, v.Enabled, v.OnCondition, v.Action, v.Value, v.MetadataKey, v.MetadataValue)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("construct insert actions sql: %v", err)
	}

	if tx != nil {
		_, err = tx.tx.Exec(sql, args...)
	} else {
		_, err = s.db.Exec(sql, args...)
	}

	if err != nil {
		return getDatabaseError(err, s, "insert rule actions")
	}
	return nil
}

func (s *RuleStore) addConditionsToRule(tx *tx, ruleId int, conditions []*models.RuleCondition) error {
	query := s.sq.Insert("rule_conditions").
		Columns("rule_id", "enabled", "case_insensitive", "inverted_match", "condition_type",
			"is_regex", "value", "date_fmt", "metadata_key", "metadata_value")

	for _, v := range conditions {
		query = query.Values(ruleId, v.Enabled, v.CaseInsensitive, v.Inverted, v.ConditionType, v.IsRegex, v.Value, v.DateFmt,
			v.MetadataKey, v.MetadataValue)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("construct insert conditions sql: %v", err)
	}

	if tx != nil {
		_, err = tx.tx.Exec(sql, args...)
	} else {
		_, err = s.db.Exec(sql, args...)

	}
	if err != nil {
		return getDatabaseError(err, s, "insert rule conditions")
	}
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
