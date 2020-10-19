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
func (s *DocumentStore) GetDocuments(userId int, paging Paging) (*[]models.Document, int, error) {
	sql := `
SELECT *
FROM documents
WHERE user_id = $1
OFFSET $2
LIMIT $3;
`

	dest := &[]models.Document{}
	err := s.db.Select(dest, sql, userId, paging.Offset, paging.Limit)

	sql = `
SELECT count(id) 
FROM documents
WHERE user_id = $1
`
	var count int
	err = s.db.Get(&count, sql, userId)
	return dest, count, err
}

// GetDocument returns document by its id. If userId != 0, user must be owner of the document.
func (s *DocumentStore) GetDocument(userId int, id int) (*models.Document, error) {
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
	return dest, err
}

// UserOwnsDocumet returns true if user has ownership for document.
func (s *DocumentStore) UserOwnsDocument(documentId, userId int) (bool, error) {

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
	return ownership, err

}

func (s *DocumentStore) GetByHash(hash string) (*models.Document, error) {

	sql := `
	SELECT id, name, filename, content, created_at, updated_at, hash, mimetype
	FROM documents
	WHERE hash = $1;
`
	object := &models.Document{}
	err := s.db.Get(object, sql, hash)
	return object, getDatabaseError(err)
}

func (s *DocumentStore) Create(doc *models.Document) error {
	sql := `
INSERT INTO documents (user_id, name, content, filename, hash, mimetype)
	VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`

	res, err := s.db.Query(sql, doc.UserId, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype)
	if err != nil {
		return getDatabaseError(err)
	}
	defer res.Close()

	if res.Next() {
		var id int
		err = res.Scan(&id)
		if err != nil {
			return getDatabaseError(err)
		} else {
			doc.Id = id
		}
	}

	return getDatabaseError(err)
}

// SetDocumentContent sets content for given document id
func (s *DocumentStore) SetDocumentContent(id int, content string) error {

	sql := `
UPDATE documents SET content=$2
WHERE id=$1;
`

	_, err := s.db.Exec(sql, id, content)
	return getDatabaseError(err)
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
	return docs, getDatabaseError(err)
}

// BulkUpdateIndexingStatus sets indexing status for all documents that are defined in ids-array.
func (s *DocumentStore) BulkUpdateIndexingStatus(indexed bool, at time.Time, ids []int) error {

	sql := `
UPDATE documents SET
awaits_indexing = $1,
indexed_at = $2,
WHERE id IN ($3);
`

	_, err := s.db.Exec(sql, !indexed, at, ids)
	return getDatabaseError(err)
}

func (s *DocumentStore) Update(doc *models.Document) error {
	doc.UpdatedAt = time.Now()
	sql := `
UPDATE documents SET 
name=$2, content=$3, filename=$4, hash=$5, mimetype=$6
updated_at=$7,
WHERE id=$1
`

	_, err := s.db.Exec(sql, doc.Id, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.UpdatedAt)
	return getDatabaseError(err)
}
