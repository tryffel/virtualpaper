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
	"github.com/jmoiron/sqlx"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataStore struct {
	db *sqlx.DB
}

func (s *MetadataStore) GetDocumentMetadata(userId int, documentId int) (*[]models.Metadata, error) {
	var sql string
	var args []interface{}

	if userId != 0 {
		sql = `
SELECT key, value
FROM metadata
LEFT JOIN documents d ON metadata.document_id = d.id
WHERE document_id = $1
AND d.user_id = $2
ORDER by key ASC;
`
		args = []interface{}{documentId, userId}

	} else {
		sql = `
SELECT key, value
FROM metadata
WHERE document_id = $1
ORDER BY key ASC;
`
		args = []interface{}{documentId}
	}

	object := &[]models.Metadata{}
	err := s.db.Select(object, sql, args...)
	return object, getDatabaseError(err)
}

func (s *MetadataStore) GetDocumentTags(userId int, documentId int) (*[]models.Tag, error) {
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
	return object, getDatabaseError(err)
}

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
		return object, len(*object), getDatabaseError(err)
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
		return object, len(*object), getDatabaseError(err)
	}
	return object, n, getDatabaseError(err)
}

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
	return object, getDatabaseError(err)
}

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
		return getDatabaseError(err)
	}

	tag.Id = id
	return nil
}
