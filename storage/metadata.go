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
	"github.com/sirupsen/logrus"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataStore struct {
	db *sqlx.DB
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
	return object, getDatabaseError(err, "metadata", "get document metadata")
}

// GetKeys returns all possible metadata-keys for user.
func (s *MetadataStore) GetKeys(userId int) (*[]models.MetadataKey, error) {
	limit := config.MaxRows
	sql := `
SELECT *
FROM metadata_keys
WHERE user_id = $1
ORDER BY key ASC
LIMIT $2;
`

	keys := &[]models.MetadataKey{}
	err := s.db.Select(keys, sql, userId, limit)
	return keys, getDatabaseError(err, "metadata", "get keys")
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
	return key, getDatabaseError(err, "metadata", "get key")
}

// GetValues returns all values to given key.
func (s *MetadataStore) GetValues(userId int, keyId int) (*[]models.MetadataValue, error) {
	limit := config.MaxRows

	sql := `
SELECT *
FROM metadata_values mv
WHERE  mv.user_id = $1
AND key_id = $2
ORDER BY value ASC
LIMIT $3;
`

	values := &[]models.MetadataValue{}
	err := s.db.Select(values, sql, userId, keyId, limit)
	return values, getDatabaseError(err, "metadata", "get key values")
}

// UpdateDocumentKeyValues updates key-values for document.
func (s *MetadataStore) UpdateDocumentKeyValues(userId int, documentId string, metadata []*models.Metadata) error {
	var sql string
	var err error

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
			return getDatabaseError(err, "metadata", "update key-values, check ownership")
		}
		if !ownership {
			return errors.ErrRecordNotFound
		}
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return getDatabaseError(err, "metadata", "update document, start tx")
	}

	sql = `
	DELETE
	FROM document_metadata m
	WHERE m.document_id = $1;
	`

	_, err = tx.Exec(sql, documentId)
	if err != nil {
		logrus.Warningf("error deleting old metadata: %v", err)
		return getDatabaseError(tx.Rollback(), "metadata", "rollback tx")
	}

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
	if err != nil {
		err = tx.Rollback()
	} else {
		err = tx.Commit()
	}
	return getDatabaseError(err, "metadata", "update document key-values")
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
	return values, getDatabaseError(err, "metadata values", "get where match_documents = true")
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
	return exists, getDatabaseError(err, "metadata", "check key-value ownership")
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
		return getDatabaseError(err, "metadata keys", "create")
	}

	if res.Next() {
		var id int
		err = res.Scan(&id)
		if err != nil {
			return getDatabaseError(err, "metadata keys", "create, scan id")
		}
		key.Id = id
	}
	return nil
}

// CreateValue creates new metadata value.
func (s *MetadataStore) CreateValue(userId int, value *models.MetadataValue) error {
	sql := `
INSERT INTO metadata_values
(user_id, key_id, value)
VALUES ($1, $2, $3)
RETURNING id;
`

	res, err := s.db.Query(sql, userId, value.KeyId, value.Value)
	if err != nil {
		return getDatabaseError(err, "metadata values", "create")
	}

	if res.Next() {
		var id int
		err = res.Scan(&id)
		if err != nil {
			return getDatabaseError(err, "metadata values", "create, scan id")
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
	return object, getDatabaseError(err, "metadata", "get document tags")
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
		return object, len(*object), getDatabaseError(err, "metadata", "get tags")
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
		return object, len(*object), getDatabaseError(err, "metadata", "scan tags count")
	}
	return object, n, getDatabaseError(err, "metadata", "get tags count")
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
	return object, getDatabaseError(err, "metadata", "get tag")
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
		return getDatabaseError(err, "metadata", "create tag")
	}

	tag.Id = id
	return nil
}
