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
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type UserStore struct {
	db    *sqlx.DB
	cache *cache.Cache
}

func newUserStore(db *sqlx.DB) *UserStore {
	store := &UserStore{
		db:    db,
		cache: cache.New(time.Minute, time.Minute),
	}
	return store
}

func (u *UserStore) getUserIdCache(id int) *models.User {
	cacheRecord, found := u.cache.Get(fmt.Sprintf("userid-%d", id))
	if found {
		user, ok := cacheRecord.(*models.User)
		if ok {
			return user
		} else {
			u.cache.Delete(fmt.Sprintf("userid-%d", id))
		}
	}
	return nil
}

func (u *UserStore) getUserNameCache(username string) *models.User {
	cacheRecord, found := u.cache.Get(fmt.Sprintf("username-%s", username))
	if found {
		user, ok := cacheRecord.(*models.User)
		if ok {
			return user
		} else {
			u.cache.Delete(fmt.Sprintf("username-%s", username))
		}
	}
	return nil
}

func (u *UserStore) setUserCache(user *models.User) {
	if user == nil || user.Id == 0 {
		return
	}
	u.cache.Set(fmt.Sprintf("userid-%d", user.Id), user, cache.DefaultExpiration)
	u.cache.Set(fmt.Sprintf("username-%s", user.Name), user, cache.DefaultExpiration)
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
VALUES ($1, $2, $3, $4) RETURNING id;

`
	rows, err := u.db.Query(sql, user.Name, "", time.Now(), user.Password)
	if err != nil {
		return getDatabaseError(err, "user", "add")
	}

	if rows.Next() {
		id := 0
		err := rows.Scan(&id)
		if err != nil {
			return getDatabaseError(err, "user", "add user, scan id")
		}
		user.Id = id
	} else {
		return errors.New("no id returned")
	}
	return nil
}

// GetUsers returns all users.
func (u *UserStore) GetUsers() (*[]models.User, error) {
	sql := `
	SELECT *
	FROM users
	LIMIT 1000;
	`

	users := &[]models.User{}
	err := u.db.Select(users, sql)
	return users, getDatabaseError(err, "users", "get many")
}

// GetUser returns single user with id.
func (u *UserStore) GetUser(userid int) (*models.User, error) {
	user := u.getUserIdCache(userid)
	if user != nil {
		return user, nil
	}

	sql := `
	SELECT *
	FROM users
	WHERE id = $1;
	`

	user = &models.User{}
	err := u.db.Get(user, sql, userid)
	if err != nil {
		return user, getDatabaseError(err, "users", "get by id")
	}

	u.setUserCache(user)
	return user, nil
}

// GetUserByName returns user matching username
func (u *UserStore) GetUserByName(username string) (*models.User, error) {
	sql := `
	SELECT *
	FROM users
	WHERE name = $1;
	`

	user := &models.User{}
	err := u.db.Get(user, sql, username)
	return user, getDatabaseError(err, "users", "get by name")
}

// Update existing user. Username cannot be changed,
func (u *UserStore) Update(user *models.User) error {
	if user.Id < 1 {
		e := errors.ErrInvalid
		e.ErrMsg = fmt.Sprintf("user (%d) does not exist", user.Id)
		return e

	}
	user.Update()

	sql := `
UPDATE users SET
email=$2, updated_at=$3, password=$4, active=$5, admin=$6
where id = $1
`

	_, err := u.db.Exec(sql, user.Id, user.Email, user.UpdatedAt, user.Password, user.IsActive, user.IsAdmin)
	return getDatabaseError(err, "user", "update")
}

func (u *UserStore) GetUserPreferences(userid int) (*models.UserPreferences, error) {

	sql := `
SELECT
       u.id AS user_id,
       u.name AS username,
       u.admin AS is_admin,
       count(d.id) AS documents_count,
       sum(d.size) AS documents_size
FROM users u
LEFT JOIN documents d ON u.id = d.user_id

WHERE u.id = $1
GROUP BY(u.id);`

	pref := &models.UserPreferences{}

	err := u.db.Get(pref, sql, userid)
	if err != nil {
		return pref, getDatabaseError(err, "user", "get preferences")
	}

	stopWords, err := u.GetPreferenceValue(userid, PreferenceStopWords)
	if err != nil {
		return pref, fmt.Errorf("get stopwords: %v", err)
	}
	if stopWords != "" {
		err = json.Unmarshal([]byte(stopWords), &pref.StopWords)
		if err != nil {
			return pref, fmt.Errorf("unmarshal stopwords: %v", err)
		}
	}

	synonyms, err := u.GetPreferenceValue(userid, PreferenceSynonyms)
	if err != nil {
		return pref, fmt.Errorf("get synonyms: %v", err)
	}
	if synonyms != "" {
		err = json.Unmarshal([]byte(synonyms), &pref.Synonyms)
		if err != nil {
			return pref, fmt.Errorf("unmarshal synonyms: %v", err)
		}
	}
	return pref, err

}

type PreferenceKey string

const (
	PreferenceStopWords PreferenceKey = "stop_words"
	PreferenceSynonyms  PreferenceKey = "synonyms"
)

func (u *UserStore) GetPreferenceValue(userId int, key PreferenceKey) (string, error) {
	sql := `
SELECT value
FROM user_preferences
WHERE user_id=$1
AND key=$2
`

	value := ""
	err := u.db.Get(&value, sql, userId, string(key))
	return value, getDatabaseErrorIgnoreEmpty(err, "preferences", "get")
}

func (u *UserStore) SetPreferenceValue(userId int, key PreferenceKey, value string) error {
	now := time.Now()

	sql := `
INSERT INTO user_preferences (user_id, "key", "value", updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, key)  DO
UPDATE SET "value"=$3, updated_at=$4
WHERE user_preferences.user_id=$1
AND user_preferences.key=$2;
`
	_, err := u.db.Exec(sql, userId, key, value, now)
	return getDatabaseErrorIgnoreEmpty(err, "preferences", "set")
}

func (u *UserStore) UpdatePreferences(userId int, stopWords []string, synonyms [][]string) error {

	stopWordsB, err := json.Marshal(stopWords)
	if err != nil {
		return fmt.Errorf("serialize stopwords: %v", err)
	}
	synonymsB, err := json.Marshal(synonyms)
	if err != nil {
		return fmt.Errorf("serialize synonyms: %v", err)
	}

	err = u.SetPreferenceValue(userId, PreferenceStopWords, string(stopWordsB))
	if err != nil {
		return fmt.Errorf("save stopwords: %v", err)
	}

	err = u.SetPreferenceValue(userId, PreferenceSynonyms, string(synonymsB))
	if err != nil {
		return fmt.Errorf("save synonyms: %v", err)
	}
	return nil
}
