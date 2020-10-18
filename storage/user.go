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
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

type UserStore struct {
	db *sqlx.DB
}

// TryLogin tries to log user in by password. Return userId = -1 and error if login fails.
func (u *UserStore) TryLogin(username, password string) (int, error) {
	sql :=
		`
SELECT id, name, password
FROM users
WHERE name = $1;
`

	user := &models.User{}
	err := u.db.Get(user, sql, username)
	if err != nil {
		return -1, err
	}

	if user.Name != username {
		return -1, fmt.Errorf("user not found")
	}

	if ok, _ := user.PasswordMatches(password); ok {
		return user.Id, nil
	}

	return -1, fmt.Errorf("user not found")
}

// AddUser adds user. Id is updated.
func (u *UserStore) AddUser(user *models.User) error {

	sql := `
INSERT INTO users (name, email, updated_at, password)
VALUES ($1, $2, $3, $4);

`
	_, err := u.db.Exec(sql, user.Name, "", time.Now(), user.Password)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserStore) GetUsers() (*[]models.User, error) {
	sql := `
	SELECT *
	FROM users
	LIMIT 1000;
	`

	users := &[]models.User{}
	err := u.db.Select(users, sql)
	return users, getDatabaseError(err)
}
