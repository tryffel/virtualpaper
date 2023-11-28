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
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"regexp"
)

// SortKey contains sortable key and order. Order 'false' = ASC, 'true' = DESC.
type SortKey struct {
	Key             string
	Order           bool
	CaseInsensitive bool
}

func NewSortKey(key string, defaultKey string, order bool, caseInsensitive bool) SortKey {
	sort := SortKey{
		Key:             key,
		Order:           order,
		CaseInsensitive: caseInsensitive,
	}

	sort.Validate(defaultKey)
	return sort
}

func (s *SortKey) SetDefaults(key string, order bool) {
	if s.Key == "" {
		s.Key = key
		s.Order = order
	}
}

func (s SortKey) SortOrder() string {
	if s.Order {
		return "DESC"
	}
	return "ASC"
}

func (s SortKey) QueryKey() string {
	if s.CaseInsensitive {
		return fmt.Sprintf("lower(%s)", s.Key)
	}
	return s.Key
}

var legalSortKey = regexp.MustCompile("([a-z_.]{0,100})")

// Validate validates sort keys and enforces the key to be legal.
func (s *SortKey) Validate(defaultKey string) {

	if legalSortKey.Match([]byte(s.Key)) {
		return
	}

	logrus.Infof("illegal sort parameter %s", s.Key)
	s.Key = defaultKey
}

type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
}

// Resource is a generic persistence storage for single resource type.
type Resource interface {
	Name() string

	parseError(e error, action string) error
}

type resource struct {
	name string
	db   *sqlx.DB
}

func (r *resource) Name() string {
	return r.name
}

func (r *resource) parseError(e error, action string) error {
	return getDatabaseError(e, r, action)
}

type tx struct {
	tx       *sqlx.Tx
	ok       bool
	resource Resource
}

func (r *resource) beginTx() (*tx, error) {
	xTx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin tx: %v", err)
	}
	tx := &tx{
		tx:       xTx,
		ok:       false,
		resource: r,
	}
	return tx, nil
}

func (tx *tx) Close() {
	var err error
	if tx.ok {
		err = tx.tx.Commit()
		if err != nil {
			err = getDatabaseError(err, tx.resource, "commit")
		}
	} else {
		err := tx.tx.Rollback()
		if err != nil {
			err = getDatabaseError(err, tx.resource, "rollback")
		}
	}
	if err != nil {
		logrus.Errorf("close transaction: %v", err)
	}
}
