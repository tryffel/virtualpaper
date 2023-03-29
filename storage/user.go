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
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type UserStore struct {
	db    *sqlx.DB
	cache *cache.Cache
}

func (s *UserStore) FlushCache() {
	s.cache.Flush()
}

func (s *UserStore) Name() string {
	return "Users"
}

func (s *UserStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func newUserStore(db *sqlx.DB) *UserStore {
	store := &UserStore{
		db:    db,
		cache: cache.New(time.Minute, time.Minute),
	}
	return store
}

func (s *UserStore) getUserIdCache(id int) *models.User {
	cacheRecord, found := s.cache.Get(fmt.Sprintf("userid-%d", id))
	if found {
		user, ok := cacheRecord.(*models.User)
		if ok {
			return user
		} else {
			s.cache.Delete(fmt.Sprintf("userid-%d", id))
		}
	}
	return nil
}

func (s *UserStore) getUserNameCache(username string) *models.User {
	cacheRecord, found := s.cache.Get(fmt.Sprintf("username-%s", username))
	if found {
		user, ok := cacheRecord.(*models.User)
		if ok {
			return user
		} else {
			s.cache.Delete(fmt.Sprintf("username-%s", username))
		}
	}
	return nil
}

func (s *UserStore) setUserCache(user *models.User) {
	if user == nil || user.Id == 0 {
		return
	}
	s.cache.Set(fmt.Sprintf("userid-%d", user.Id), user, cache.DefaultExpiration)
	s.cache.Set(fmt.Sprintf("username-%s", user.Name), user, cache.DefaultExpiration)
}

// TryLogin tries to log user in by password. Return userId = -1 and error if login fails.
func (s *UserStore) TryLogin(username, password string) (int, error) {
	sql :=
		`
SELECT id, name, password
FROM users
WHERE LOWER(name) =$1
AND active = TRUE;
`

	user := &models.User{}
	err := s.db.Get(user, sql, strings.ToLower(username))
	if err != nil {
		return -1, s.parseError(err, "get user by username")
	}

	if strings.ToLower(user.Name) != strings.ToLower(username) {
		return -1, fmt.Errorf("user not found")
	}

	if ok, _ := user.PasswordMatches(password); ok {
		return user.Id, nil
	}

	return -1, fmt.Errorf("user not found")
}

// AddUser adds user. Id is updated.
func (s *UserStore) AddUser(user *models.User) error {

	sql := `
INSERT INTO users (name, email, updated_at, password, active, admin)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;
`

	var email interface{}
	if user.Email == "" {
		email = nil
	} else {
		email = user.Email
	}

	rows, err := s.db.Query(sql, user.Name, email, time.Now(), user.Password, user.IsActive, user.IsAdmin)
	if err != nil {
		return s.parseError(err, "add")
	}

	if rows.Next() {
		id := 0
		err := rows.Scan(&id)
		if err != nil {
			return s.parseError(err, "add user, scan id")
		}
		user.Id = id
	} else {
		return errors.New("no id returned")
	}

	err = s.UpdatePreferences(user.Id, []string{}, [][]string{})
	return err
}

// GetUsers returns all users.
func (s *UserStore) GetUsers() (*[]models.User, error) {
	sql := `
SELECT 
	id,
	name,
	COALESCE(email, '') AS email, 
	password,
	active,
	admin,
	created_at,
	updated_at
FROM users
LIMIT 1000;
	`

	users := &[]models.User{}
	err := s.db.Select(users, sql)
	return users, s.parseError(err, "get users")
}

func (s *UserStore) GetUsersInfo() (*[]models.UserInfo, error) {

	sql := `
SELECT 
	u.id as user_id, 
	u.name as username, 
	COALESCE(email, '') AS email, 
	active, 
	admin, 
	u.created_at as created_at, 
	u.updated_at as updated_at, 
	count(d) as documents_count, 
	sum(d."size") as documents_size 
from users u 
left join documents d on u.id = d.user_id 
group by u.id
order by u.name asc
limit 1000;`

	info := &[]models.UserInfo{}
	err := s.db.Select(info, sql)
	return info, s.parseError(err, "get detailed user info")
}

// GetUser returns single user with id.
func (s *UserStore) GetUser(userid int) (*models.User, error) {
	user := s.getUserIdCache(userid)
	if user != nil {
		return user, nil
	}

	sql := `
SELECT 
	id,
	name,
	COALESCE(email, '') AS email, 
	password,
	active,
	admin,
	created_at,
	updated_at
FROM users
WHERE id = $1;
	`

	user = &models.User{}
	err := s.db.Get(user, sql, userid)
	if err != nil {
		return user, s.parseError(err, "get by id")
	}

	s.setUserCache(user)
	return user, nil
}

// GetUserByName returns user matching username
func (s *UserStore) GetUserByName(username string) (*models.User, error) {
	sql := `
SELECT 
	id,
	name,
	COALESCE(email, '') AS email, 
	password,
	active,
	admin,
	created_at,
	updated_at
FROM users
WHERE name = $1;
	`

	user := &models.User{}
	err := s.db.Get(user, sql, username)
	return user, s.parseError(err, "get by name")
}

func (s *UserStore) GetUserByEmail(email string) (*models.User, error) {
	sql := `
SELECT
   	id,
	name,
	COALESCE(email, '') AS email, 
	password,
	active,
	admin,
	created_at,
	updated_at
FROM users
WHERE email = $1`

	user := &models.User{}
	err := s.db.Get(user, sql, email)
	return user, s.parseError(err, "get by email")
}

// Update existing user. Username cannot be changed,
func (s *UserStore) Update(user *models.User) error {
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

	var email interface{}
	if user.Email == "" {
		email = nil
	} else {
		email = user.Email
	}

	_, err := s.db.Exec(sql, user.Id, email, user.UpdatedAt, user.Password, user.IsActive, user.IsAdmin)
	if err == nil {
		s.setUserCache(user)
	}
	return s.parseError(err, "update")
}

func (s *UserStore) GetUserPreferences(userid int) (*models.UserPreferences, error) {

	sql := `
SELECT
       s.id AS user_id,
       s.name AS username,
       s.admin AS is_admin,
       count(d.id) AS documents_count,
       sum(d.size) AS documents_size
FROM users s
LEFT JOIN documents d ON s.id = d.user_id
WHERE s.id = $1
GROUP BY(s.id);`

	pref := &models.UserPreferences{}

	err := s.db.Get(pref, sql, userid)
	if err != nil {
		return pref, s.parseError(err, "get preferences")
	}

	stopWords, err := s.GetPreferenceValue(userid, PreferenceStopWords)
	if err != nil {
		return pref, fmt.Errorf("get stopwords: %v", err)
	}
	if stopWords != "" {
		err = json.Unmarshal([]byte(stopWords), &pref.StopWords)
		if err != nil {
			return pref, fmt.Errorf("unmarshal stopwords: %v", err)
		}
	}

	synonyms, err := s.GetPreferenceValue(userid, PreferenceSynonyms)
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

func (s *UserStore) GetPreferenceValue(userId int, key PreferenceKey) (string, error) {
	sql := `
SELECT value
FROM user_preferences
WHERE user_id=$1
AND key=$2
`

	value := ""
	err := s.db.Get(&value, sql, userId, string(key))
	return value, s.parseError(err, "get preference value")
}

func (s *UserStore) SetPreferenceValue(userId int, key PreferenceKey, value string) error {
	now := time.Now()

	sql := `
INSERT INTO user_preferences (user_id, "key", "value", updated_at)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, key)  DO
UPDATE SET "value"=$3, updated_at=$4
WHERE user_preferences.user_id=$1
AND user_preferences.key=$2;
`
	_, err := s.db.Exec(sql, userId, key, value, now)
	return s.parseError(err, "set preference value")
}

func (s *UserStore) UpdatePreferences(userId int, stopWords []string, synonyms [][]string) error {

	stopWordsB, err := json.Marshal(stopWords)
	if err != nil {
		return fmt.Errorf("serialize stopwords: %v", err)
	}
	synonymsB, err := json.Marshal(synonyms)
	if err != nil {
		return fmt.Errorf("serialize synonyms: %v", err)
	}

	err = s.SetPreferenceValue(userId, PreferenceStopWords, string(stopWordsB))
	if err != nil {
		return fmt.Errorf("save stopwords: %v", err)
	}

	err = s.SetPreferenceValue(userId, PreferenceSynonyms, string(synonymsB))
	if err != nil {
		return fmt.Errorf("save synonyms: %v", err)
	}
	return nil
}

func (s *UserStore) AddPasswordResetToken(token *models.PasswordResetToken) error {
	err := token.Validate()
	if err != nil {
		return err
	}

	sql := `INSERT INTO password_reset_tokens (token, user_id, created_at, updated_at, expires_at) 
VALUES ($1, $2, $3, $4, $5) RETURNING id`

	id := 0
	err = s.db.Get(&id, sql, token.Token, token.UserId, token.CreatedAt, token.UpdatedAt, token.ExpiresAt)
	if err == nil {
		token.Id = id
	}
	return s.parseError(err, "save password reset token")
}

func (s *UserStore) GetPasswordResetTokenByHash(tokenId int) (*models.PasswordResetToken, error) {
	sql := `SELECT * FROM password_reset_tokens WHERE id=$1`

	token := &models.PasswordResetToken{}
	err := s.db.Get(token, sql, tokenId)
	return token, s.parseError(err, "get by id")
}

func (s *UserStore) DeletePasswordResetToken(tokenId int) error {
	sql := `DELETE FROM password_reset_tokens WHERE id = $1`
	_, err := s.db.Exec(sql, tokenId)
	return s.parseError(err, "delete password reset token")
}

func (s *UserStore) DeleteExpiredPasswordResetTokens() (int, error) {
	sql := `DELETE FROM password_reset_tokens WHERE expires_at < now()`
	out, err := s.db.Exec(sql)
	if err != nil {
		return 0, s.parseError(err, "delete expired tokens")
	}

	affected, err := out.RowsAffected()
	if err != nil {
		logrus.Warningf("get rows affected for deleting expired tokens: %v", err)
	}
	return int(affected), nil
}
