package models

import (
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
	"time"
)

type Token struct {
	Timestamp
	Id            int       `json:"id" db:"id"`
	UserId        int       `json:"user_id" db:"user_id"`
	Key           string    `json:"key" db:"key"`
	Name          string    `json:"name" db:"name"`
	IpAddr        string    `json:"ip_addr" db:"ip_address"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	LastSeen      time.Time `json:"last_seen" db:"last_seen"`
	LastConfirmed time.Time `json:"last_confirmed" db:"last_confirmed"`
}

func (t *Token) Init() error {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	var err error
	t.Key, err = uuid.GenerateUUID()
	if err == nil {
		return nil
	}
	logrus.Warningf("failed to generate uuid: %v, retrying", err)
	t.Key, err = uuid.GenerateUUID()
	if err == nil {
		return fmt.Errorf("generate key: %v", err)
	}
	return nil
}

func (t *Token) HasExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return t.ExpiresAt.Before(time.Now())
}

func (t *Token) ConfirmationExpired() bool {
	if t.LastConfirmed.IsZero() {
		return true
	}
	return t.LastConfirmed.Add(time.Minute * 15).Before(time.Now())
}
