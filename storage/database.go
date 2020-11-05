package storage

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"tryffel.net/go/virtualpaper/config"
)

type Database struct {
	conn *sqlx.DB

	UserStore     *UserStore
	DocumentStore *DocumentStore
	JobStore      *JobStore
	MetadataStore *MetadataStore
}

func NewDatabase() (*Database, error) {
	db := &Database{}

	conf := config.C.Database

	url := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Database)
	var err error
	db.conn, err = sqlx.Connect("postgres", url)

	if err != nil {
		return db, err
	}

	db.UserStore = newUserStore(db.conn)
	db.DocumentStore = &DocumentStore{db: db.conn}
	db.JobStore = &JobStore{db: db.conn}
	db.MetadataStore = &MetadataStore{db: db.conn}
	return db, nil
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
