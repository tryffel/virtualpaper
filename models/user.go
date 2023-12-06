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

package models

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Timestamp
	Id       int    `db:"id"`
	Name     string `db:"name"`
	Password string `db:"password"`
	Email    string `db:"email"`
	IsAdmin  bool   `db:"admin"`
	IsActive bool   `db:"active"`
}

func (u *User) SetPassword(newPassw string) error {
	if len(newPassw) < 8 {
		return errors.New("password must be minimum of 8 characters")
	} else if len(newPassw) > 150 {
		return errors.New("password must be maximum of 150 characters")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(newPassw), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *User) PasswordMatches(password string) (bool, error) {
	if u.Password == "" {
		return false, fmt.Errorf("password not set")
	}
	currentPass := []byte(u.Password)
	err := bcrypt.CompareHashAndPassword(currentPass, []byte(password))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}

	return false, err
}

// UserPreferences are per-user preferences and configuration options.
type UserPreferences struct {
	UserId        int       `json:"user_id" db:"user_id"`
	UserName      string    `json:"user_name" db:"username"`
	Email         string    `json:"email" db:"email"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	DocumentCount Int       `json:"documents_count" db:"documents_count"`
	DocumentsSize Int       `json:"documents_size" db:"documents_size"`
	IsAdmin       bool      `json:"is_admin" db:"is_admin"`
}

type UserInfo struct {
	UserId        int       `json:"id" db:"user_id"`
	UserName      string    `json:"user_name" db:"username"`
	Email         string    `json:"email" db:"email"`
	IsActive      bool      `json:"is_active" db:"active"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	DocumentCount Int       `json:"documents_count" db:"documents_count"`
	DocumentsSize Int       `json:"documents_size" db:"documents_size"`
	IsAdmin       bool      `json:"is_admin" db:"admin"`
	LastSeen      time.Time `json:"last_seen" db:"last_seen"`

	Indexing              bool `json:"indexing"`
	TotalDocumentsIndexed int  `json:"documents_indexed_count"`
}

type PasswordResetToken struct {
	Timestamp
	Id        int       `db:"id"`
	UserId    int       `db:"user_id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (p *PasswordResetToken) HasExpired() bool {
	now := time.Now()
	return now.After(p.ExpiresAt)
}

func (p *PasswordResetToken) Validate() error {
	if p.UserId == 0 {
		return fmt.Errorf("no userid")
	}
	if len(p.Token) < 20 {
		return fmt.Errorf("token is too short")
	}
	if p.HasExpired() {
		return fmt.Errorf("token has expired")
	}
	return nil
}

func (p *PasswordResetToken) TokenMatches(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty token")
	}
	if p.Token == "" {
		return false, fmt.Errorf("password not set")
	}
	hash := []byte(p.Token)
	err := bcrypt.CompareHashAndPassword(hash, []byte(token))
	if err == nil {
		return true, nil
	}

	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	}
	return false, err
}
