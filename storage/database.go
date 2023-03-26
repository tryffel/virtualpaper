package storage

import (
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
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
