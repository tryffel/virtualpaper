package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"tryffel.net/go/virtualpaper/errors"
)

type DocumentSharePermission struct {
	UserId      int         `db:"user_id"`
	Username    string      `db:"user_name"`
	DocumentId  string      `db:"document_id"`
	Permissions Permissions `db:"permissions"`
	Timestamp
}

type Permissions struct {
	Read   bool `json:"read"`
	Write  bool `json:"write"`
	Delete bool `json:"delete"`
}

func (p Permissions) Value() (driver.Value, error) {
	out, err := json.Marshal(p)
	return string(out), err
}

func (p *Permissions) Scan(src interface{}) error {
	if src == nil {
		p.Read = false
		p.Write = false
		p.Delete = false
		return nil
	}
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

type UpdateUserSharing struct {
	UserId      int         `json:"user_id"`
	Permissions Permissions `json:"permissions"`
}
