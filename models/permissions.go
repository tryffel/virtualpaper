package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"tryffel.net/go/virtualpaper/errors"
)

type DocumentSharePermission struct {
	UserId      int         `db:"user_id"`
	DocumentId  string      `db:"document_id"`
	Permissions Permissions `db:"permissions"`
	Timestamp
}

type Permissions struct {
	Read   bool
	Write  bool
	Delete bool
}

func (p Permissions) Value() (driver.Value, error) {
	out, err := json.Marshal(p)
	return string(out), err
}

func (p *Permissions) Scan(src interface{}) error {
	array, ok := src.([]byte)
	if !ok {
		return errors.New("source not array")
	}
	err := json.Unmarshal(array, p)
	if err != nil {
		return fmt.Errorf("json: %v", err)
	}
	return nil
}
