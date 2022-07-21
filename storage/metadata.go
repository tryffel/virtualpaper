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
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataStore struct {
	db    *sqlx.DB
	sq    squirrel.StatementBuilderType
	cache *cache.Cache
}

func NewMetadataStore(db *sqlx.DB) *MetadataStore {
	return &MetadataStore{
		db:    db,
		sq:    squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		cache: cache.New(time.Second*30, time.Second*30),
	}
}
func (m *MetadataStore) cacheNameUserKeys(userId int) string {
	return fmt.Sprintf("user-%d-keys", userId)
}

func (m *MetadataStore) cacheNameUserKeyValues(userId int, key string) string {
	return fmt.Sprintf("user-%d-key-%s-values", userId, key)
}

func (m *MetadataStore) getCachedKeys(userId int) *[]models.MetadataKey {
	data, ok := m.cache.Get(m.cacheNameUserKeys(userId))
	if !ok {
		return nil
	}

	keys, ok := data.(*[]models.MetadataKey)
	if !ok {
		m.flushCachedUserKeys(userId)
		return nil
	}
	return keys
}

func (m *MetadataStore) flushCachedUserKeys(userId int) {
	m.cache.Delete(m.cacheNameUserKeys(userId))
}

func (m *MetadataStore) getCachedKeyValues(userId int, key string) *[]models.Metadata {
	data, ok := m.cache.Get(m.cacheNameUserKeyValues(userId, key))
	if !ok {
		return nil
	}

	keys, ok := data.(*[]models.Metadata)
	if !ok {
		m.flushCachedUserKeyValues(userId, key)
		return nil
	}
	return keys
}

func (m *MetadataStore) flushCachedUserKeyValues(userId int, key string) {
	m.cache.Delete(m.cacheNameUserKeyValues(userId, key))
}

func (s MetadataStore) Name() string {
	return "Metadata"
}

func (s MetadataStore) parseError(err error, action string) error {
	return getDatabaseError(err, s, action)
}

// GetDocumentMetadata returns key-value metadata for given document. If userId != 0, user must own document.
func (s *MetadataStore) GetDocumentMetadata(userId int, documentId string) (*[]models.Metadata, error) {
	var sql string
	var args []interface{}

	if userId != 0 {
		sql = `

SELECT
	mk.id AS key_id,
	mk.key AS key,
	mv.id AS value_id,
	mv.value AS value
FROM documents d
LEFT JOIN document_metadata dm ON d.id = dm.document_id
LEFT JOIN metadata_keys mk ON dm.key_id = mk.id
LEFT JOIN metadata_values mv ON dm.value_id = mv.id
WHERE d.id = $1
AND d.user_id = $2
ORDER BY key ASC;
`
		args = []interface{}{documentId, userId}

	} else {
		sql = `
SELECT
	mk.id AS key_id,
	mk.key AS key,
	mv.id AS value_id,
	mv.value AS value
FROM documents d
LEFT JOIN document_metadata dm ON d.id = dm.document_id
LEFT JOIN metadata_keys mk ON dm.key_id = mk.id
LEFT JOIN metadata_values mv ON dm.value_id = mv.id
WHERE d.id = $1
ORDER BY key ASC;
`
		args = []interface{}{documentId}
	}

	object := &[]models.Metadata{}
	err := s.db.Select(object, sql, args...)
	if err != nil {
		if strings.Contains(err.Error(), "converting NULL to int") {
			logrus.Debugf("got empty row from metadata query: doc %s, %v", documentId, err)
			return object, nil
			// no rows returned
		}
	}
	return object, s.parseError(err, "get document metadata")
}

func (s *MetadataStore) GetUserKeysCached(userId int) (*[]models.MetadataKey, error) {

	keys := s.getCachedKeys(userId)
	if keys != nil {
		return keys, nil
	}

	query := s.sq.Select("mk.id as id", "lower(mk.key) as key", "mk.comment as comment",
		"mk.created_at as created_at", "COUNT(dm.document_id) as documents_count").
		From("metadata_keys mk").LeftJoin("document_metadata dm ON mk.id = dm.key_id").
		Where(squirrel.Eq{"user_id": userId}).GroupBy("mk.id").
		OrderBy("documents_count DESC").Limit(config.MaxRows)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("construct sql: %v", err)
	}

	keys = &[]models.MetadataKey{}
	err = s.db.Select(keys, sql, args...)

	if err != nil {
		return keys, s.parseError(err, "get keys")
	}

	s.cache.SetDefault(s.cacheNameUserKeys(userId), keys)
	return keys, nil
}
func (s *MetadataStore) GetUserKeyValuesCached(userId int, key string) (*[]models.Metadata, error) {
	values := s.getCachedKeyValues(userId, key)
	if values != nil {
		return values, nil
	}

	query := s.sq.Select(
		"mv.id as value_id",
		"mk.id as key_id",
		"lower(mv.value) as value",
		"lower(mk.key) as key").
		From("metadata_values mv").
		LeftJoin("metadata_keys mk on mv.key_id = mk.id").
		LeftJoin("document_metadata dm on mv.id = dm.value_id").
		Where(squirrel.Eq{"mv.user_id": userId}).
		Where(squirrel.Eq{"lower(mk.key)": key}).
		GroupBy("mv.id", "mv.value", "mk.id", "mk.key").
		OrderBy("count(dm.document_id) DESC").Limit(config.MaxRows)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("construct sql: %v", err)
	}

	values = &[]models.Metadata{}
	err = s.db.Select(values, sql, args...)
	if err != nil {
		return values, s.parseError(err, "get key values")
	}

	s.cache.SetDefault(s.cacheNameUserKeyValues(userId, key), values)
	return values, nil
}

// GetKeys returns all possible metadata-keys for user.
func (s *MetadataStore) GetKeys(userId int, ids []int, sort SortKey, paging Paging) (*[]models.MetadataKey, error) {
	paging.Validate()
	sort.Validate("id")
	query := s.sq.Select("mk.id as id", "mk.key as key", "mk.comment as comment",
		"mk.created_at as created_at", "COUNT(dm.document_id) as documents_count").
		From("metadata_keys mk").LeftJoin("document_metadata dm ON mk.id = dm.key_id").
		Where(squirrel.Eq{"user_id": userId}).GroupBy("mk.id")

	if len(ids) > 0 {
		query = query.Where(squirrel.Eq{"id": ids})
	}

	query = query.Limit(uint64(paging.Limit)).Offset(uint64(paging.Offset))
	query = query.OrderBy(sort.QueryKey() + " " + sort.SortOrder())
	keys := &[]models.MetadataKey{}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("construct sql: %v", err)
	}
	err = s.db.Select(keys, sql, args...)
	return keys, s.parseError(err, "get keys")
}

// GetKeys returns all possible metadata-keys for user.
func (s *MetadataStore) GetKey(userId int, keyId int) (*models.MetadataKey, error) {
	sql := `
SELECT *
FROM metadata_keys
WHERE user_id = $1
AND id = $2;
`

	key := &models.MetadataKey{}

	err := s.db.Get(key, sql, userId, keyId)
	return key, s.parseError(err, "get key")
}

// GetValues returns all values to given key.
func (s *MetadataStore) GetValues(userId int, keyId int, sort SortKey, paging Paging) (*[]models.MetadataValue, error) {
	paging.Validate()
	sort.Validate("id")
	query := s.sq.Select(
		"mv.id as id",
		"mv.value as value",
		"mk.key as key",
		"mv.created_at as created_at",
		"match_documents",
		"match_type",
		"match_filter",
		"count(dm.document_id) as documents_count").
		From("metadata_values mv").
		LeftJoin("document_metadata dm on mv.id = dm.value_id").
		LeftJoin("metadata_keys mk on mv.key_id = mk.id").
		Where(squirrel.Eq{"mv.user_id": userId}).
		Where(squirrel.Eq{"mv.key_id": keyId}).GroupBy("mv.id", "mv.value", "mk.key").
		OrderBy(sort.QueryKey() + " " + sort.SortOrder()).Limit(uint64(paging.Limit)).Offset(uint64(paging.Offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("construct sql: %v", err)
	}

	values := &[]models.MetadataValue{}
	err = s.db.Select(values, sql, args...)
	return values, s.parseError(err, "get key values")
}

// UpdateDocumentKeyValues updates key-values for document.
func (s *MetadataStore) UpdateDocumentKeyValues(userId int, documentId string, metadata []*models.Metadata) error {
	logrus.Debugf("update document %s metadata, key-values: %d", documentId, len(metadata))

	var sql string
	var err error

	if len(metadata) > 0 {

		if userId != 0 {
			sql = `
SELECT CASE WHEN EXISTS
    (
        SELECT id
        FROM documents
        WHERE id = $1
        AND user_id = $2
    )
    THEN TRUE
    ELSE FALSE
END;
`

			var ownership bool
			err = s.db.Get(&ownership, sql, documentId, userId)
			if err != nil {
				return s.parseError(err, "update key-values, check ownership")
			}
			if !ownership {
				return errors.ErrRecordNotFound
			}
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return s.parseError(err, "update document, start tx")
	}

	sql = `
	DELETE
	FROM document_metadata m
	WHERE m.document_id = $1;
	`

	_, err = tx.Exec(sql, documentId)
	if err != nil {
		logrus.Warningf("error deleting old metadata: %v", err)
		return s.parseError(tx.Rollback(), "rollback tx")
	}

	if len(metadata) > 0 {

		sql = `	
	INSERT INTO document_metadata (document_id, key_id, value_id)
	VALUES `

		var args []interface{}
		args = append(args, documentId)
		for i, v := range metadata {
			if i > 0 {
				sql += ", "
			}
			value := fmt.Sprintf("($1, $%d, $%d)", i*2+2, i*2+3)
			sql += value
			args = append(args, v.KeyId, v.ValueId)
		}

		_, err = tx.Exec(sql, args...)

	}
	if err != nil {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}
	return s.parseError(err, "update document key-values")
}

// GetUserValuesWithMatching retusn all metadata values that
// have Metadatavalue.MatchDocuments enabled.
func (s *MetadataStore) GetUserValuesWithMatching(userId int) (*[]models.MetadataValue, error) {
	sql := `
SELECT * FROM
metadata_values
WHERE user_id = $1
AND match_documents = TRUE;
`

	values := &[]models.MetadataValue{}
	err := s.db.Select(values, sql, userId)
	return values, s.parseError(err, "(value) get where match_documents = true")
}

// KeyValuePairExists checks whether given pair actually exists and is user owns them.
func (s *MetadataStore) KeyValuePairExists(userId, key, value int) (bool, error) {

	sql := `
SELECT CASE WHEN EXISTS
(
	select id
   	m metadata_values
    	e user_id = $1
    	d key_id = $2
    	d id = $3
)
THEN TRUE
ELSE FALSE
END AS exists;
`

	exists := false
	err := s.db.Get(&exists, sql, userId, key, value)
	return exists, s.parseError(err, "check key-value ownership")
}

// CreateKey creates new metadata key.
func (s *MetadataStore) CreateKey(userId int, key *models.MetadataKey) error {

	sql := `
INSERT INTO metadata_keys
(user_id, key, comment)
VALUES ($1, $2, $3)
RETURNING id;
`

	res, err := s.db.Query(sql, userId, key.Key, key.Comment)
	if err != nil {
		return s.parseError(err, "create key")
	}

	if res.Next() {
		var id int
		err = res.Scan(&id)
		if err != nil {
			return s.parseError(err, "create key, scan id")
		}
		key.Id = id
	}
	s.flushCachedUserKeys(userId)
	return nil
}

// CreateValue creates new metadata value.
func (s *MetadataStore) CreateValue(userId int, value *models.MetadataValue) error {
	sql := `
INSERT INTO metadata_values
(user_id, key_id, value, match_documents, match_type, match_filter)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`

	res, err := s.db.Query(sql, userId, value.KeyId, value.Value, value.MatchDocuments, value.MatchType, value.MatchFilter)
	if err != nil {
		return s.parseError(err, "create value")
	}

	if res.Next() {
		var id int
		err = res.Scan(&id)
		if err != nil {
			return s.parseError(err, "create value, scan id")
		}
		value.Id = id
	}
	return nil
}

// GetDocumentTags returns tags for given document.
func (s *MetadataStore) GetDocumentTags(userId int, documentId string) (*[]models.Tag, error) {
	sql := `
select tags.id as id, tags.key as key, tags.comment as comment
from tags
LEFT JOIN document_tags dt on tags.id = dt.tag_id
LEFT JOIN documents d on dt.document_id = d.id
WHERE dt.document_id= $1
and d.user_id = $2
order by key asc
limit 100;
`
	object := &[]models.Tag{}
	err := s.db.Select(object, sql, documentId, userId)
	return object, s.parseError(err, "get document tags")
}

// GetTags returns all tags for user.
func (s *MetadataStore) GetTags(userid int, paging Paging) (*[]models.TagComposite, int, error) {

	sql := `
SELECT
       tags.id AS id, COUNT(tags.id) AS document_count,
       tags.key AS key, tags.comment AS comment,
       tags.created_at AS created_at, tags.updated_at AS updated_at
FROM tags
LEFT JOIN document_tags dt on tags.id = dt.tag_id
WHERE tags.user_id = $1
GROUP BY (tags.id)
ORDER BY tags.key asc
OFFSET $2
LIMIT $3;
`

	object := &[]models.TagComposite{}

	err := s.db.Select(object, sql, userid, paging.Offset, paging.Limit)
	if err != nil {
		return object, len(*object), s.parseError(err, "get tags")
	}

	sql = `SELECT
	COUNT(tags.id) AS count
	FROM tags
	WHERE tags.user_id = $1;
	`

	n := 0
	row := s.db.QueryRow(sql, userid)
	err = row.Scan(&n)
	if err != nil {
		return object, len(*object), s.parseError(err, "scan tags count")
	}
	return object, n, s.parseError(err, "get tags count")
}

// GetTag returns tag with given id.
func (s *MetadataStore) GetTag(userId, tagId int) (*models.TagComposite, error) {
	sql := `
SELECT 
	tags.id AS id, 
	tags.key AS key, 
	tags.comment AS comment, 
	COUNT(d.id) as document_count, 
	tags.created_at AS created_at, 
	tags.updated_at AS updated_at
FROM tags
LEFT JOIN document_tags dt ON tags.id = dt.tag_id
LEFT JOIN documents d ON dt.document_id = d.id
WHERE tags.id = $1
AND tags.user_id = $2
GROUP BY (tags.id);
`

	object := &models.TagComposite{}
	err := s.db.Get(object, sql, tagId, userId)
	return object, s.parseError(err, "get tag")
}

// CreateTag creates new tag.
func (s *MetadataStore) CreateTag(userId int, tag *models.Tag) error {
	sql := `
INSERT INTO tags (user_id, key, comment, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
`

	tag.CreatedAt = time.Now()
	tag.UpdatedAt = time.Now()

	id := 0
	row := s.db.QueryRow(sql, userId, tag.Key, tag.Comment, tag.CreatedAt, tag.UpdatedAt)
	err := row.Scan(&id)
	if err != nil {
		return s.parseError(err, "create tag")
	}

	tag.Id = id
	return nil
}

func (s *MetadataStore) UserHasKeyValue(userId, keyId, valueId int) (bool, error) {

	sql := `
SELECT CASE WHEN EXISTS (
    SELECT mv.id
        FROM metadata_values mv
        LEFT JOIN metadata_keys mk ON mv.key_id = mk.id
        WHERE mk.user_id = $1
            AND mv.id = $3
    AND mk.id = $2
    )
THEN TRUE ELSE FALSE END AS exists;
`

	var ownership bool
	err := s.db.Get(&ownership, sql, userId, keyId, valueId)
	return ownership, s.parseError(err, "check user has key-value")
}

func (s *MetadataStore) UserHasKey(userId, keyId int) (bool, error) {
	sql := `
SELECT CASE WHEN EXISTS (
    SELECT mk.id
    FROM metadata_keys mk 
    WHERE mk.user_id=$1
      AND mk.id=$2
    )  
THEN TRUE ELSE FALSE END AS exists;
`
	var ownership bool
	err := s.db.Get(&ownership, sql, userId, keyId)
	return ownership, s.parseError(err, "check user has key")
}

func (s *MetadataStore) UserHasKeys(userId int, keys []int) (bool, error) {

	sql := `
SELECT count(distinct(id)) 
FROM metadata_keys
WHERE user_id=$1 AND id IN (
`
	args := make([]interface{}, len(keys)+1)
	args[0] = fmt.Sprintf("%d", userId)
	for i, v := range keys {
		if i > 0 {
			sql += ","
		}
		sql += fmt.Sprintf("$%d", i+2)
		args[i+1] = v
	}
	sql += ");"
	var keyCount int
	err := s.db.Get(&keyCount, sql, args...)
	if err != nil {
		return false, s.parseError(err, "check user owns metadata keys")
	}
	return keyCount == len(keys), s.parseError(err, "check user owns metadata")
}

func (s *MetadataStore) UpdateValue(value *models.MetadataValue) error {
	sql := `
	UPDATE metadata_values
	SET value=$1, match_documents=$2, match_type=$3, match_filter=$4
	WHERE id=$5;
`

	_, err := s.db.Exec(sql, value.Value, value.MatchDocuments, value.MatchType, value.MatchFilter, value.Id)
	return s.parseError(err, "update value")
}

func (s *MetadataStore) UpdateKey(key *models.MetadataKey) error {
	sql := `
UPDATE metadata_keys 
SET key=$1, comment=$2
WHERE id=$3;
`

	_, err := s.db.Exec(sql, key.Key, key.Comment, key.Id)
	if err != nil {
		return s.parseError(err, "update key")
	}
	s.flushCachedUserKeys(key.UserId)
	return nil
}

// CheckKeyValuesExist verifies key-value pairs exist and user owns them.
func (s *MetadataStore) CheckKeyValuesExist(userId int, values []models.Metadata) error {
	array := make(squirrel.Or, len(values))
	for i, key := range values {
		array[i] = squirrel.And{squirrel.Eq{"key_id": key.KeyId}, squirrel.Eq{"id": key.ValueId}}
	}

	query := s.sq.Select("count(id)").From("metadata_values").
		Where(squirrel.And{squirrel.Eq{"user_id": userId}, array})
	sql, args, err := query.ToSql()
	if err != nil {
		err := errors.ErrInternalError
		err.Err = err
		return err
	}
	var count int
	err = s.db.Get(&count, sql, args...)
	if err != nil {
		return getDatabaseError(err, s, "verify metadata exists")
	}

	if count == len(values) {
		return nil
	}

	userErr := errors.ErrInvalid
	userErr.ErrMsg = "metadata does not exist"
	return userErr
}

func (s *MetadataStore) UpsertDocumentMetadata(userId int, documents []string, metadata []*models.Metadata) error {
	// when checking metadata: need to remove duplicate keys

	sql := `
INSERT INTO document_metadata (document_id, key_id, value_id) VALUES %s 
ON CONFLICT (document_id, key_id, value_id) DO NOTHING
`

	sqlParams := ""

	index := 1
	args := make([]interface{}, 0, len(documents)*len(metadata))
	for iDoc, vDoc := range documents {
		if iDoc > 0 {
			sqlParams += ","
		}
		docIndex := index
		args = append(args, vDoc)
		for i, v := range metadata {
			if i > 0 {
				sqlParams += ","
			}
			sqlParams += fmt.Sprintf("($%d, $%d, $%d)", docIndex, index+1, index+2)
			args = append(args, v.KeyId, v.ValueId)
			index += 2
		}
		index += 1
	}

	sql = fmt.Sprintf(sql, sqlParams)
	_, err := s.db.Exec(sql, args...)
	return s.parseError(err, "upsert multiple documents metadata")
}

func (s *MetadataStore) DeleteDocumentsMetadata(userId int, documents []string, metadata []*models.Metadata) error {

	sqlFormat := `
DELETE FROM document_metadata 
WHERE 
	document_id IN (%s) 
	AND key_id IN (%s) 
	AND value_id IN (%s);
		`

	args := make([]interface{}, 0, len(documents)+len(metadata))
	docArgs := ""
	keyArgs := ""
	valueArgs := ""

	index := 0

	for i, v := range documents {
		if i > 0 {
			docArgs += ","
		}
		docArgs += fmt.Sprintf("$%d", i+1)
		args = append(args, v)
	}
	index = len(documents)

	for i, v := range metadata {
		if i > 0 {
			keyArgs += ","
			valueArgs += ","
		}

		keyArgs += fmt.Sprintf("$%d", index+1)
		valueArgs += fmt.Sprintf("$%d", index+2)

		args = append(args, v.KeyId, v.ValueId)
		index += 2
	}
	sql := fmt.Sprintf(sqlFormat, docArgs, keyArgs, valueArgs)
	_, err := s.db.Exec(sql, args...)
	return s.parseError(err, "remove multiple documents metadata")
}

// DeleteKey deletes metadata key.
// If userId != 0, user has to own the key.
// This will cascade the deletion to any table that uses metadata keys too: document_metadata, rules.
func (s *MetadataStore) DeleteKey(userId int, keyId int) error {
	query := s.sq.Delete("metadata_keys").Where("id=?", keyId)
	if userId != 0 {
		query = query.Where("user_id=?", userId)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		e := errors.ErrInternalError
		e.ErrMsg = "bad sql"
		e.Err = err
		return e
	}

	_, err = s.db.Exec(sql, args...)
	if err != nil {
		return s.parseError(err, "delete key")
	}
	s.flushCachedUserKeys(userId)
	return nil
}

// DeleteValue deletes metadata value from key.
// If userId != 0, user has to own the value.
// This will cascade the deletion to any table that uses metadata keys too: document_metadata, rules.
func (s *MetadataStore) DeleteValue(userId int, valueId int) error {
	query := s.sq.Delete("metadata_values").Where("id=?", valueId)
	if userId != 0 {
		query = query.Where("user_id=?", userId)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		e := errors.ErrInternalError
		e.ErrMsg = "bad sql"
		e.Err = err
		return e
	}

	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "delete value")
}
