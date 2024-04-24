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

package migration

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

var Migrations = []Migrator{
	&Migration{
		Name:   "initial schema",
		Level:  1,
		Schema: schemaV1,
	},
	&Migration{
		Name:   "file processing queue",
		Level:  2,
		Schema: schemaV2,
	},
	&Migration{
		Name:   "document metadata",
		Level:  3,
		Schema: schemaV3,
	},
	&Migration{
		Name:   "processing rules",
		Level:  4,
		Schema: schemaV4,
	},
	&Migration{
		Name:   "support for user preferences",
		Level:  5,
		Schema: schemaV5,
	},
	&Migration{
		Name:   "cascade deletions, add deleted_at col to documents",
		Level:  6,
		Schema: schemaV6,
	},
	&Migration{
		Name:   "add rules v2",
		Level:  7,
		Schema: schemaV7,
	},
	&Migration{
		Name:   "add missing on cascade delete constraints",
		Level:  8,
		Schema: schemaV8,
	},
	&Migration{
		Name:   "add document history table",
		Level:  9,
		Schema: schemaV9,
	},
	&Migration{
		Name:   "add linked documents table",
		Level:  10,
		Schema: schemaV10,
	},
	&Migration{
		Name:   "add document show history table",
		Level:  11,
		Schema: schemaV11,
	},
	&Migration{
		Name:   "add table for password resets",
		Level:  12,
		Schema: schemaV12,
	},
	&Migration{
		Name:   "add support for persisted auth tokens",
		Level:  13,
		Schema: schemaV13,
	},
	&Migration{
		Name:   "auth_tokens store last_confirmed status",
		Level:  14,
		Schema: schemaV14,
	},
	&Migration{
		Name:   "enforce case-insensitive unique user names and emails",
		Level:  15,
		Schema: schemaV15,
	},
	&Migration{
		Name:   "split process_queue.step to action and action_order",
		Level:  16,
		Schema: schemaV16,
	},
	&Migration{
		Name:   "add list of languages",
		Level:  17,
		Schema: schemaV17,
	},
	&Migration{
		Name:   "add support for metadata icon and style",
		Level:  18,
		Schema: schemaV18,
	},
	&Migration{
		Name:   "add user_shared_documents table",
		Level:  19,
		Schema: schemaV19,
	},
	&Migration{
		Name:   "add favorite column to documents",
		Level:  20,
		Schema: schemaV20,
	},
	&Migration{
		Name:   "add rule trigger column",
		Level:  21,
		Schema: schemaV21,
	},
	&Migration{
		Name:   "add document property tables",
		Level:  22,
		Schema: schemaV22,
	},
}

type Schema struct {
	Level     int       `db:"level"`
	Success   int       `db:"success"`
	Timestamp time.Time `db:"timestamp"`
	TookMs    int       `db:"took_ms"`
}

// Migrator describes single migration level
type Migrator interface {
	// Get migration name
	MName() string
	// Get migration level
	MLevel() int
	// Get valid sql string to execute
	MSchema() string
}

// Migration implements migrator
type Migration struct {
	Name   string
	Level  int
	Schema string
}

func (m *Migration) MName() string {
	return m.Name
}

func (m *Migration) MLevel() int {
	return m.Level
}

func (m *Migration) MSchema() string {
	return m.Schema
}

// Migrate runs given migrations
func Migrate(db *sqlx.DB, migrations []Migrator) error {
	current, err := CurrentVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %v", err)
	}

	if current.Level == 0 {
		_, err := db.Exec(`
CREATE TABLE "schemas" (
	"level"	INTEGER,
	"success"	INTEGER NOT NULL,
	"timestamp"	TIMESTAMP NOT NULL,
	"took_ms"	INTEGER NOT NULL,
	PRIMARY KEY("level")
);
`)
		if err != nil {
			return fmt.Errorf("failed create schema table: %v", err)
		}

	} else {
		if current.Success == 0 {
			return fmt.Errorf("previous migration has failed")
		}
	}

	if current.Level == migrations[len(migrations)-1].MLevel() {
		logrus.Debug("No new migrations to run")
		return nil
	}

	lastLevel := current.Level

	if lastLevel > len(migrations) {
		return fmt.Errorf("schema level newer than supported by this version: got %d, expected %d",
			lastLevel, migrations[len(migrations)-1].MLevel())
	}

	for _, v := range migrations[current.Level:] {
		logrus.Warningf("Migrating database schema %d -> %d", lastLevel, v.MLevel())
		err := migrateSingle(db, v)
		if err != nil {
			return fmt.Errorf("failed to run migrations: %v", err)
		}
		lastLevel = v.MLevel()
	}
	logrus.Warning("Migrations ok")
	return nil
}

// Run single migration
func migrateSingle(db *sqlx.DB, migration Migrator) error {
	start := time.Now()

	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("begin transaction: %v", err)
	}
	defer tx.Rollback()

	_, merr := tx.Exec(migration.MSchema())
	if merr != nil {
		return fmt.Errorf("migration failed: insert schema: %v", merr)
	}

	s := &Schema{
		Level:     migration.MLevel(),
		Timestamp: time.Now(),
		TookMs:    int(time.Since(start).Nanoseconds() / 1000000),
	}

	if merr == nil {
		s.Success = 1
	} else {
		s.Success = 0
	}

	_, err = tx.Exec("INSERT INTO schemas (level, success, timestamp, took_ms) "+
		"VALUES ($1, $2, $3, $4)", s.Level, s.Success, s.Timestamp, s.TookMs)

	if err != nil {
		return fmt.Errorf("insert to 'schemas': %v", err)
	}
	return tx.Commit()
}

// CurrentVersion returns current version
func CurrentVersion(db *sqlx.DB) (Schema, error) {
	current := Schema{}
	err := db.Get(&current, "SELECT * FROM schemas ORDER BY level DESC LIMIT 1")

	if err != nil {
		e := err.Error()

		if strings.HasSuffix(e, "relation \"schemas\" does not exist") ||
			strings.HasSuffix(e, "no such table: schemas") {
			return Schema{
				Level:     0,
				Success:   0,
				Timestamp: time.Time{},
				TookMs:    0,
			}, nil
		}

		return Schema{}, fmt.Errorf("failed to query schema: %v", err)

	}
	return current, nil
}
