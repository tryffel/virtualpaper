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
	"golang.org/x/crypto/bcrypt"
	"time"
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
	UserId    int       `json:"user_id" db:"user_id"`
	UserName  string    `json:"user_name" db:"username"`
	Email     string    `json:"email" db:"email"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
