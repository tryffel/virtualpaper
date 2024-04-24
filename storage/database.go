package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"tryffel.net/go/virtualpaper/config"
)

// Database connects to postgresql database
// and contains store for each model/relation.
type Database struct {
	conn *sqlx.DB

	UserStore     *UserStore
	DocumentStore *DocumentStore
	JobStore      *JobStore
	MetadataStore *MetadataStore
	StatsStore    *StatsStore
	RuleStore     *RuleStore
	AuthStore     *AuthStore
	PropertyStore *PropertyStore
}

func (d *Database) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.conn.Exec(query, args...)
}

func (d *Database) ExecSq(sql squirrel.Sqlizer) (sql.Result, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql: %v", err)
	}
	return d.conn.Exec(query, args...)
}

func (d *Database) QuerySq(sql squirrel.Sqlizer) (*sqlx.Rows, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql: %v", err)
	}
	return d.conn.Queryx(query, args...)
}

func (d *Database) SelectSq(destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("generate sql: %v", err)
	}
	return d.conn.Select(destination, query, args...)
}

func (d *Database) GetSq(destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("generate sql: %v", err)
	}
	return d.conn.Get(destination, query, args...)
}

func (d *Database) Get(destination interface{}, query string, args ...interface{}) error {
	return d.conn.Get(destination, query, args...)
}

func (d *Database) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.conn.ExecContext(ctx, query, args...)
}

func (d *Database) ExecContextSq(ctx context.Context, sql squirrel.Sqlizer) (sql.Result, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql: %v", err)
	}
	return d.conn.ExecContext(ctx, query, args...)
}

func (d *Database) QueryContextSq(ctx context.Context, sql squirrel.Sqlizer) (*sqlx.Rows, error) {
	query, args, err := sql.ToSql()
	if err != nil {
		return nil, fmt.Errorf("generate sql: %v", err)
	}
	return d.conn.QueryxContext(ctx, query, args...)
}

func (d *Database) SelectContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	return d.conn.SelectContext(ctx, destination, query, args...)
}

func (d *Database) Select(destination interface{}, sql string, args ...interface{}) error {
	return d.conn.Select(destination, sql, args...)
}

func (d *Database) GetContextSq(ctx context.Context, destination interface{}, sql squirrel.Sqlizer) error {
	query, args, err := sql.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	return d.conn.GetContext(ctx, destination, query, args...)
}

// NewDatabase returns working instance of database connection.
func NewDatabase(conf config.Database) (*Database, error) {
	db := &Database{}

	url := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Database)

	if conf.NoSSL {
		url += " sslmode=disable"
	}

	var err error
	db.conn, err = sqlx.Connect("postgres", url)

	if err != nil {
		return db, err
	}
	db.MetadataStore = NewMetadataStore(db.conn)
	db.UserStore = newUserStore(db.conn)
	db.DocumentStore = NewDocumentStore(db.conn, db.MetadataStore)
	db.JobStore = newJobStore(db.conn)
	db.StatsStore = NewStatsStore(db.conn)
	db.RuleStore = newRuleStore(db.conn, db.MetadataStore)
	db.AuthStore = newAuthStore(db.conn)
	db.PropertyStore = NewPropertyStore(db.conn)
	return db, nil
}

// NewMockDatabase returns mock database instance
func NewMockDatabase(matcher sqlmock.QueryMatcher) (*Database, sqlmock.Sqlmock, error) {
	if matcher == nil {
		matcher = sqlmock.QueryMatcherRegexp
	}
	mockDb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		return nil, nil, fmt.Errorf("init mock db: %v", err)
	}

	db := &Database{
		conn: sqlx.NewDb(mockDb, "sqlmock"),
	}
	db.MetadataStore = NewMetadataStore(db.conn)
	db.UserStore = newUserStore(db.conn)
	db.DocumentStore = NewDocumentStore(db.conn, db.MetadataStore)
	db.JobStore = newJobStore(db.conn)
	db.StatsStore = &StatsStore{db: db.conn}
	db.AuthStore = newAuthStore(db.conn)
	db.PropertyStore = NewPropertyStore(db.conn)
	return db, mock, nil
}

func (d *Database) Close() error {
	return d.conn.Close()
}

func (d *Database) Engine() *sqlx.DB {
	return d.conn
}

type Paging struct {
	Offset int
	Limit  int
}

func (p *Paging) Validate() {
	p.Limit = config.MaxRecords(p.Limit)
}
