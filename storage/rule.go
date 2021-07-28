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
}

func (s *RuleStore) Name() string {
	return "Rules"
}

func (s *RuleStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func newRuleStore(db *sqlx.DB) *RuleStore {
	store := &RuleStore{
		db:    db,
		cache: cache.New(5*time.Minute, time.Minute),
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

/*
func (s *RuleStore) GetUserRules(userId int, paging Paging) (*[]models.Match, error) {
	sql := `
SELECT *
FROM process_rules
WHERE user_id = $1
OFFSET $2
LIMIT $3;`

	rules := &[]models.Match{}
	err := s.db.Select(rules, sql, userId, paging.Offset, paging.Limit)
	return rules, s.parseError(err, "get user rules")
}

func (s *RuleStore) GetUserRule(userId, ruleId int) (*models.Match, error) {
	sql := `
SELECT *
FROM process_rules
WHERE user_id = $1
AND id = $2;`

	rule := &models.Match{}
	err := s.db.Get(rule, sql, userId, ruleId)
	return rule, s.parseError(err, "get user rule")
}

func (s *RuleStore) AddRule(userId int, rule *models.Match) error {
	sql := `
INSERT INTO process_rules
(user_id, rule_type, filter, comment, action, active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`

	rule.Update()
	action, err := rule.Action.Value()
	if err != nil {
		return fmt.Errorf("serialize action: %v", err)
	}
	rows, err := s.db.Query(sql, userId, string(rule.Type), rule.Filter, rule.Comment, action, rule.Active)
	if err != nil {
		return s.parseError(err, "add user rule")
	}

	if rows.Next() {
		err = rows.Scan(&rule.Id)
	}

	return s.parseError(err, "add rule, scan rows")
}


*/
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

	sql = `
SELECT
    rule_conditions.id as id,
    rule_id,
    rule_conditions.enabled as enabled,
    case_insensitive,
    inverted_match,
    condition_type,
    is_regex,
    value,
    metadata_key,
    metadata_value
FROM rule_conditions
         LEFT JOIN rules on rule_conditions.rule_id = rules.id
WHERE rules.user_id = $1
  AND rule_conditions.enabled = true
order by rule_id, rule_conditions.id asc;
`

	conditions := &[]models.RuleCondition{}
	err = s.db.Select(conditions, sql, userId)
	if err != nil {
		return nil, s.parseError(err, "get rule conditions")
	}

	actions := &[]models.RuleAction{}
	sql = `
SELECT
    rule_actions.id AS id,
    rule_id,
    rule_actions.enabled AS enabled,
    on_condition,
    rule_actions.value AS value,
    rule_actions.action as action,
    metadata_key,
    metadata_value
FROM rule_actions
	LEFT JOIN rules ON rule_actions.rule_id = rules.id
WHERE rules.user_id = $1
	AND rule_actions.enabled = TRUE
ORDER BY rule_id, rule_actions.id ASC;
`

	err = s.db.Select(actions, sql, userId)
	if err != nil {
		return nil, s.parseError(err, "get rule conditions")
	}

	ruleArr := make([]*models.Rule, len(*rules))

	for i, _ := range *rules {
		rule := (*rules)[i]
		rule.Conditions = make([]*models.RuleCondition, 0, 10)
		rule.Actions = make([]*models.RuleAction, 0, 10)

		for conditionI, condition := range *conditions {
			if condition.RuleId == rule.Id {
				rule.Conditions = append(rule.Conditions, &(*conditions)[conditionI])
			}
		}

		for actionI, action := range *actions {
			if action.RuleId == rule.Id {
				rule.Actions = append(rule.Actions, &(*actions)[actionI])
			}
		}

		ruleArr[i] = &rule
	}
	return ruleArr, nil
}
