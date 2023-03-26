package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
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
		Columns("user_id", "key", "name", "expires_at", "last_seen").
		Values(token.UserId, token.Key, token.Name, token.ExpiresAt, token.LastSeen).Suffix("RETURNING id")

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

	builder := s.sq.Select("id", "user_id", "key", "name", "created_at", "updated_at", "expires_at", "last_seen").
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
		updatebuilder := s.sq.Update("auth_tokens").Set("last_seen", token.LastSeen)
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
