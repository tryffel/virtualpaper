package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

type AuthStore struct {
	db    *sqlx.DB
	cache *cache.Cache
	sq    squirrel.StatementBuilderType
}

func (s *AuthStore) FlushCache() {
	s.cache.Flush()
}

func (s *AuthStore) Name() string {
	return "Auth"
}

func (s *AuthStore) parseError(e error, action string) error {
	return getDatabaseError(e, s, action)
}

func newAuthStore(db *sqlx.DB) *AuthStore {
	store := &AuthStore{
		db:    db,
		cache: cache.New(time.Second*15, time.Minute),
		sq:    squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
	return store
}

func (s *AuthStore) getTokenKeyCache(key string) *models.Token {
	cacheRecord, found := s.cache.Get(fmt.Sprintf("token-%s", key))
	if found {
		token, ok := cacheRecord.(*models.Token)
		if ok {
			return token
		} else {
			s.cache.Delete(fmt.Sprintf("token-%s", key))
		}
	}
	return nil
}

func (s *AuthStore) deleteTokenFromCache(key string) {
	s.cache.Delete(fmt.Sprintf("token-%s", key))
}

func (s *AuthStore) setTokenCache(token *models.Token) {
	if token == nil || token.Id == 0 || token.Key == "" {
		return
	}
	s.cache.Set(fmt.Sprintf("token-%s", token.Key), token, cache.DefaultExpiration)
}

func (s *AuthStore) InsertToken(token *models.Token) error {
	builder := s.sq.Insert("auth_tokens").
		Columns("user_id", "key", "name", "expires_at", "last_seen", "ip_address", "last_confirmed").
		Values(token.UserId, token.Key, token.Name, token.ExpiresAt, token.LastSeen, token.IpAddr, token.LastConfirmed).
		Suffix("RETURNING id")

	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %v", err)
	}

	id := 0
	err = s.db.Get(&id, sql, args...)
	if err != nil {
		return s.parseError(err, "insert token")
	}
	token.Id = id
	return nil
}

func (s *AuthStore) GetToken(key string, updateLastSeen bool) (*models.Token, error) {
	cached := s.getTokenKeyCache(key)
	if cached != nil {
		return cached, nil
	}

	builder := s.sq.Select("id", "user_id", "key", "name", "created_at", "updated_at", "expires_at", "last_seen", "last_confirmed").
		From("auth_tokens").
		Where("key=?", key)

	sql, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql: %v", err)
	}

	token := &models.Token{}
	err = s.db.Get(token, sql, args...)
	if err != nil {
		return nil, s.parseError(err, "get token")
	}

	if updateLastSeen {
		token.LastSeen = time.Now()
		updatebuilder := s.sq.Update("auth_tokens").Set("last_seen", token.LastSeen).Where("key = ?", key)
		sql, args, err = updatebuilder.ToSql()
		if err != nil {
			return nil, fmt.Errorf("build sql: %v", err)
		}
		_, err = s.db.Exec(sql, args...)
		if err != nil {
			return nil, s.parseError(err, "get token")
		}
	}
	s.setTokenCache(token)
	return token, nil
}

func (s *AuthStore) UpdateTokenConfirmation(key string, confirmed time.Time) error {
	builder := s.sq.Update("auth_tokens").Set("last_confirmed", confirmed).Where("key = ?", key)
	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %v", err)
	}

	_, err = s.db.Exec(sql, args...)
	if err != nil {
		return s.parseError(err, "update token last_confirmed")
	}
	s.deleteTokenFromCache(key)
	return nil
}

func (s *AuthStore) RevokeToken(key string) error {
	s.deleteTokenFromCache(key)

	builder := s.sq.Delete("auth_tokens").Where("key=?", key)
	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("build sql: %v", err)
	}

	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "delete token")
}

func (s *AuthStore) DeleteExpiredAuthTokens() (int, error) {
	// expires_at must be non-zero value and expired
	sql := `DELETE FROM auth_tokens WHERE expires_at < now() AND EXTRACT(EPOCH from expires_at) > 1`
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
