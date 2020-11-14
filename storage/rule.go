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

func (s *RuleStore) GetUserRules(userId int, paging Paging) (*[]models.Rule, error) {
	sql := `
SELECT * 
FROM process_rules
WHERE user_id = $1
OFFSET $2
LIMIT $3;`

	rules := &[]models.Rule{}
	err := s.db.Select(rules, sql, userId, paging.Offset, paging.Limit)
	return rules, getDatabaseError(err, "rules", "get user rules")
}

func (s *RuleStore) AddRule(userId int, rule *models.Rule) error {
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
		return getDatabaseError(err, "rules", "get user rules")
	}

	if rows.Next() {
		err = rows.Scan(&rule.Id)
	}

	return getDatabaseError(err, "rules", "get user rules")
}

// GetActiveUresRules returns all active rules (with some limit) for given user.
func (s *RuleStore) GetActiveUserRules(userId int) (*[]models.Rule, error) {

	sql := `
SELECT * 
FROM process_rules
WHERE user_id = $1
LIMIT $2;`

	rules := &[]models.Rule{}
	err := s.db.Select(rules, sql, userId, config.MaxRulesToProcess)
	return rules, getDatabaseError(err, "get active user rules", "")
}
