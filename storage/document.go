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

type DocumentStore struct {
	db *sqlx.DB
}

// GetDocuments returns user's documents according to paging. In addition, return total count of documents available.
func (s *DocumentStore) GetDocuments(userId int, paging Paging, sort SortKey, limitContent bool) (*[]models.Document, int, error) {
	sort.SetDefaults("date", false)

	var contenSelect string
	if limitContent {
		contenSelect = "LEFT(content, 500) as content"
	} else {
		contenSelect = "content"
	}

	sql := `
SELECT id, name, ` + contenSelect + `, filename, created_at, updated_at
hash, mimetype, size, date, description
FROM documents
WHERE user_id = $1
ORDER BY ` + sort.Key + " " + sort.SortOrder() + `
OFFSET $2
LIMIT $3;
`

	dest := &[]models.Document{}
	err := s.db.Select(dest, sql, userId, paging.Offset, paging.Limit)
	if err != nil {
		return dest, 0, getDatabaseError(err, "documents", "get")
	}

	if limitContent && len(*dest) > 0 {
		for i, _ := range *dest {
			if len((*dest)[i].Content) > 499 {
				(*dest)[i].Content += "..."
			}
		}
	}

	sql = `
SELECT count(id) 
FROM documents
WHERE user_id = $1
`
	var count int
	err = s.db.Get(&count, sql, userId)
	err = getDatabaseError(err, "documents", "get documents")
	return dest, count, err
}

// GetDocument returns document by its id. If userId != 0, user must be owner of the document.
func (s *DocumentStore) GetDocument(userId int, id string) (*models.Document, error) {
	sql := `
SELECT *
FROM documents
WHERE id = $1
`

	args := []interface{}{id}
	if userId != 0 {
		sql += " AND user_id = $2;"
		args = append(args, userId)
	}
	dest := &models.Document{}
	err := s.db.Get(dest, sql, args...)
	return dest, getDatabaseError(err, "documents", "get document")
}

// UserOwnsDocumet returns true if user has ownership for document.
func (s *DocumentStore) UserOwnsDocument(documentId string, userId int) (bool, error) {

	sql := `
select case when exists
    (
        select id
        from documents
        where id = $1
        and user_id = $2
    )
    then true
    else false
end;
`

	var ownership bool

	err := s.db.Get(&ownership, sql, documentId, userId)
	return ownership, getDatabaseError(err, "documents", "check ownership")
}

func (s *DocumentStore) GetByHash(hash string) (*models.Document, error) {

	sql := `
	SELECT id, name, filename, content, created_at, updated_at, hash, mimetype, description
	FROM documents
	WHERE hash = $1;
`
	object := &models.Document{}
	err := s.db.Get(object, sql, hash)
	return object, getDatabaseError(err, "documents", "get by hash")
}

func (s *DocumentStore) Create(doc *models.Document) error {
	sql := `
INSERT INTO documents (user_id, name, content, filename, hash, mimetype, size, description)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`

	doc.Init()

	_, err := s.db.Exec(sql, doc.UserId, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Description)
	return getDatabaseError(err, "documents", "created")
}

// SetDocumentContent sets content for given document id
func (s *DocumentStore) SetDocumentContent(id string, content string) error {

	sql := `
UPDATE documents SET content=$2
WHERE id=$1;
`

	_, err := s.db.Exec(sql, id, content)
	return getDatabaseError(err, "documents", "set content")
}

// GetContent returns full content. If userId != 0, user must own the document of given id.
func (s *DocumentStore) GetContent(userId int, id string) (*string, error) {
	sql := `
SELECT content
FROM documents
WHERE id=$1
`
	if userId != 0 {
		sql += " AND user_id=$2"
	}

	content := ""
	var err error
	if userId != 0 {
		err = s.db.Get(&content, sql, id, userId)
	} else {
		err = s.db.Get(&content, sql, id, userId)
	}
	return &content, getDatabaseError(err, "documents", "get content")
}

// GetNeedsIndexing returns documents that need indexing ( awaits_indexing=true). If userId != 0, return
// only documents for that user, else return any documents.
func (s *DocumentStore) GetNeedsIndexing(userId int, paging Paging) (*[]models.Document, error) {

	sql := `
SELECT * 
FROM documents
WHERE awaits_indexing=True
`

	args := []interface{}{paging.Offset, paging.Limit}
	if userId != 0 {
		sql += " AND user_id = $3 "
		args = append(args, userId)
	}

	sql += "OFFSET $1 LIMIT $2;"

	docs := &[]models.Document{}

	err := s.db.Select(docs, sql, args...)
	return docs, getDatabaseError(err, "documents", "get needs indexing")
}

// BulkUpdateIndexingStatus sets indexing status for all documents that are defined in ids-array.
func (s *DocumentStore) BulkUpdateIndexingStatus(indexed bool, at time.Time, ids []string) error {

	sql := `
UPDATE documents SET
awaits_indexing = $1,
indexed_at = $2,
WHERE id IN ($3);
`

	_, err := s.db.Exec(sql, !indexed, at, ids)
	return getDatabaseError(err, "documents", "bulk update indexing")
}

// Update sets complete document record, not just changed attributes. Thus document must be read before updating.
func (s *DocumentStore) Update(doc *models.Document) error {
	doc.UpdatedAt = time.Now()
	sql := `
UPDATE documents SET 
name=$2, content=$3, filename=$4, hash=$5, mimetype=$6, size=$7, date=$8,
updated_at=$9, description=$10
WHERE id=$1
`

	_, err := s.db.Exec(sql, doc.Id, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Date, doc.UpdatedAt, doc.Description)
	return getDatabaseError(err, "documents", "update")
}
