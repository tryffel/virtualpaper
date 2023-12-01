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
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
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
	Exec(query string, args ...interface{}) (sql.Result, error)
}
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	SelectContext(ctx context.Context, destination interface{}, query string, args ...interface{}) error
}

type ExecerSq interface {
	ExecContextSq(ctx context.Context, sql squirrel.Sqlizer) (sql.Result, error)
	ExecSq(sql squirrel.Sqlizer) (sql.Result, error)
}
type QuerierSq interface {
	QueryContextSq(ctx context.Context, sql squirrel.Sqlizer) (*sqlx.Rows, error)
	QuerySq(sql squirrel.Sqlizer) (*sqlx.Rows, error)
	SelectContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error
	SelectSq(destination interface{}, sql squirrel.Sqlizer) error
	GetContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error
	GetSq(destination interface{}, sql squirrel.Sqlizer) error
	Get(destination interface{}, query string, args ...interface{}) error
}

type SqlExecer interface {
	Execer
	ExecerSq
	ExecerSq
	QuerierSq
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
	context  context.Context
}

func NewTx(db *Database, ctx context.Context) (*tx, error) {
	transaction, err := db.conn.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %v", err)
	}
	return &tx{
		tx:       transaction,
		ok:       false,
		resource: nil,
		context:  ctx,
	}, nil
}

func (tx *tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(ctx, query, args...)
}

func (tx *tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.tx.ExecContext(tx.context, query, args...)
}

func (tx *tx) ExecSq(sql squirrel.Sqlizer) (sql.Result, error) {
	return tx.ExecContextSq(tx.context, sql)
}

func (tx *tx) ExecContextSq(ctx context.Context, sql squirrel.Sqlizer) (sql.Result, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql: %v", err)
	}
	return tx.tx.ExecContext(ctx, query, args...)
}

func (tx *tx) QueryContextSq(ctx context.Context, sql squirrel.Sqlizer) (*sqlx.Rows, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql: %v", err)
	}
	return tx.tx.QueryxContext(ctx, query, args...)
}

func (tx *tx) QuerySq(sql squirrel.Sqlizer) (*sqlx.Rows, error) {
	return tx.QueryContextSq(tx.context, sql)
}

func (tx *tx) SelectContext(ctx context.Context, destination interface{}, query string, args ...interface{}) error {
	return tx.tx.SelectContext(ctx, destination, query, args...)
}

func (tx *tx) SelectContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	return tx.tx.SelectContext(ctx, destination, query, args...)
}
func (tx *tx) SelectSq(destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	return tx.tx.SelectContext(tx.context, destination, query, args...)
}

func (tx *tx) GetContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	return tx.tx.GetContext(ctx, destination, query, args...)
}

func (tx *tx) GetSq(destination interface{}, sql squirrel.Sqlizer) error {
	return tx.GetContextSq(tx.context, destination, sql)
}

func (tx *tx) Get(destination interface{}, query string, args ...interface{}) error {
	return tx.tx.GetContext(tx.context, destination, query, args...)
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

/*
func (tx *tx) CloseSilent() {
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

*/

func (tx *tx) Commit() error {
	var err error
	err = tx.tx.Commit()
	if err != nil {
		return getDatabaseError(err, tx.resource, "commit")
	}
	return nil
}

func (tx *tx) Close() {
	err := tx.tx.Rollback()
	if err != nil {
		if !errors.Is(err, sql.ErrTxDone) {
			logrus.Error(getDatabaseError(err, tx.resource, "rollback"))
		}
	}
}
