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
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

const UserIdInternal = -1

type DocumentStore struct {
	db            *sqlx.DB
	sq            squirrel.StatementBuilderType
	metadataStore *MetadataStore
}

func NewDocumentStore(db *sqlx.DB, mt *MetadataStore) *DocumentStore {
	return &DocumentStore{
		db:            db,
		sq:            squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		metadataStore: mt,
	}
}

func (s DocumentStore) Name() string {
	return "Documents"
}

func (s DocumentStore) parseError(err error, action string) error {
	return getDatabaseError(err, s, action)
}

// GetDocuments returns user's documents according to paging. In addition, return total count of documents available.
func (s *DocumentStore) GetDocuments(userId int, paging Paging, sort SortKey, limitContent bool, showTrash bool) (*[]models.Document, int, error) {
	sort.SetDefaults("date", false)

	var contenSelect string
	if limitContent {
		contenSelect = "LEFT(content, 500) as content"
	} else {
		contenSelect = "content"
	}

	sql := `
SELECT id, name, ` + contenSelect + `, 
	filename, created_at, updated_at,
	hash, mimetype, size, date, description, lang, deleted_at
FROM documents
WHERE user_id = $1 AND deleted_at %s
ORDER BY ` + sort.QueryKey() + " " + sort.SortOrder() + `
OFFSET $2
LIMIT $3
`

	var deletedAt = "IS NULL"
	if showTrash {
		deletedAt = "IS NOT NULL"
	}

	sql = fmt.Sprintf(sql, deletedAt)

	dest := &[]models.Document{}
	err := s.db.Select(dest, sql, userId, paging.Offset, paging.Limit)
	if err != nil {
		return dest, 0, s.parseError(err, "get")
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
WHERE user_id = $1 AND deleted_at 
` + deletedAt
	var count int
	err = s.db.Get(&count, sql, userId)
	err = s.parseError(err, "get documents")
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
	return dest, s.parseError(err, "get document")
}

// GetDocument returns document by its id. If userId != 0, user must be owner of the document.
func (s *DocumentStore) GetDocumentsById(userId int, id []string) (*[]models.Document, error) {

	query := s.sq.Select("*").From("documents").Where(squirrel.Eq{"id": id})
	if userId != 0 {
		query = query.Where("user_id = ?", userId)
	}
	query = query.OrderBy("id ASC")
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("sql: %v", err)
	}
	dest := &[]models.Document{}
	err = s.db.Select(dest, sql, args...)
	return dest, s.parseError(err, "get document")
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
	return ownership, s.parseError(err, "check ownership")

}

func (s *DocumentStore) UserOwnsDocuments(userId int, documents []string) (bool, error) {
	sql := `SELECT count(distinct(id)) FROM documents
	WHERE user_id=$1 AND id IN (
	`

	args := make([]interface{}, len(documents)+1)
	args[0] = fmt.Sprintf("%d", userId)
	for i, v := range documents {
		if i > 0 {
			sql += ","
		}
		sql += fmt.Sprintf("$%d", i+2)
		args[i+1] = v
	}
	sql += ");"
	var documentCount int
	err := s.db.Get(&documentCount, sql, args...)

	if err != nil {
		return false, s.parseError(err, "check user owns documents")
	}

	return documentCount == len(documents), s.parseError(err, "check user owns documents")
}

// GetByHash returns a document by its hash.
// If userId != 0, user has to be the owner of the document.
func (s *DocumentStore) GetByHash(userId int, hash string) (*models.Document, error) {

	sql := `
	SELECT *
	FROM documents
	WHERE hash = $1
`
	args := []interface{}{hash}
	if userId != 0 {
		sql += " AND user_id = $2"
		args = append(args, userId)
	}

	object := &models.Document{}
	err := s.db.Get(object, sql, args...)
	if err != nil {
		e := s.parseError(err, "get by hash")
		if errors.Is(e, errors.ErrRecordNotFound) {
			return object, nil
		}
		return object, e
	}
	return object, nil
}

func (s *DocumentStore) Create(doc *models.Document) error {
	sql := `
INSERT INTO documents (id, user_id, name, content, filename, hash, mimetype, size, description, date, lang)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id;`

	doc.Init()

	rows, err := s.db.Query(sql, doc.Id, doc.UserId, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Description, doc.Date, doc.Lang)
	if err != nil {
		return s.parseError(err, "created")
	}

	if rows.Next() {
		err := rows.Scan(&doc.Id)
		if err != nil {
			return fmt.Errorf("scan new document id: %v", err)
		}
		rows.Close()
	}
	err = addDocumentHistoryAction(s.db, s.sq, []models.DocumentHistory{{DocumentId: doc.Id, Action: models.DocumentHistoryActionCreate, OldValue: "", NewValue: doc.Name}}, doc.UserId)
	s.sq.Select()
	return err
}

func addDocumentHistoryAction(db *sqlx.DB, queryBuilder squirrel.StatementBuilderType, items []models.DocumentHistory, userId int) error {
	if len(items) == 0 {
		return nil
	}

	query := queryBuilder.Insert("document_history").Columns("document_id", "action", "old_value", "new_value")
	if userId != UserIdInternal {
		query = query.Columns("user_id")
	}

	if userId != UserIdInternal {
		for _, v := range items {
			query = query.Values(v.DocumentId, v.Action, v.OldValue, v.NewValue, userId)
		}
	} else {
		for _, v := range items {
			query = query.Values(v.DocumentId, v.Action, v.OldValue, v.NewValue)
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("create sql: %v", err)
	}
	_, err = db.Exec(sql, args...)
	return getDatabaseError(err, &DocumentStore{}, "add document_history actions")
}

// SetDocumentContent sets content for given document id
func (s *DocumentStore) SetDocumentContent(id string, content string) error {

	sql := `
UPDATE documents SET content=$2
WHERE id=$1;
`

	_, err := s.db.Exec(sql, id, content)
	return s.parseError(err, "set content")
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
	return &content, s.parseError(err, "get content")
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
	return docs, s.parseError(err, "get needs indexing")
}

// Update sets complete document record, not just changed attributes. Thus document must be read before updating.
func (s *DocumentStore) Update(userId int, doc *models.Document) error {

	// TODO: metadata diff is not saved, bc the document metadata is saved separately
	oldDoc, err := s.GetDocument(0, doc.Id)
	if err != nil {
		return s.parseError(err, "get document by id")
	}
	doc.UpdatedAt = time.Now()
	sql := `
UPDATE documents SET 
name=$2, content=$3, filename=$4, hash=$5, mimetype=$6, size=$7, date=$8,
updated_at=$9, description=$10, lang=$11
WHERE id=$1
`

	_, err = s.db.Exec(sql, doc.Id, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Date, doc.UpdatedAt, doc.Description, doc.Lang)
	if err != nil {
		return s.parseError(err, "update")
	}

	diff, err := oldDoc.Diff(doc, userId)
	if err != nil {
		return fmt.Errorf("get diff for document: %v", err)
	}

	err = addDocumentHistoryAction(s.db, s.sq, diff, userId)
	logrus.Infof("User %d edited document %s with %d actions", userId, doc.Id, len(diff))
	return err
}

func (s *DocumentStore) SetModifiedAt(docIds []string, modifiedAt time.Time) error {
	query := s.sq.Update("documents").Set("updated_at", modifiedAt).Where(squirrel.Eq{"id": docIds})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "update modified_at")
}

func (s *DocumentStore) MarkDocumentDeleted(userId int, docId string) error {
	query := s.sq.Update("documents").Set("deleted_at", time.Now()).Where("id=?", docId)
	if userId != 0 {
		query = query.Where("user_id = ?", userId)
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	if err != nil {
		return s.parseError(err, "mark document deleted")
	}

	err = addDocumentHistoryAction(s.db, s.sq, []models.DocumentHistory{{
		DocumentId: docId,
		Action:     models.DocumentHistoryActionDelete,
		OldValue:   "",
		NewValue:   "",
		UserId:     userId,
	},
	}, userId)
	if err != nil {
		return fmt.Errorf("add history entry: %v", err)
	}
	return nil
}

func (s *DocumentStore) MarkDocumentNonDeleted(userId int, docId string) error {
	query := s.sq.Update("documents").Set("deleted_at", nil).Where("id=?", docId)
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	if err != nil {
		return s.parseError(err, "mark document deleted")
	}
	err = addDocumentHistoryAction(s.db, s.sq, []models.DocumentHistory{{
		DocumentId: docId,
		Action:     models.DocumentHistoryActionRestore,
		OldValue:   "",
		NewValue:   "",
		UserId:     userId,
	},
	}, userId)
	if err != nil {
		return fmt.Errorf("add history entry: %v", err)
	}
	return nil
}

func (s *DocumentStore) DeleteDocument(docId string) error {
	sql := `DELETE FROM documents WHERE id = $1`
	_, err := s.db.Exec(sql, docId)
	return s.parseError(err, "delete")
}

func (s *DocumentStore) GetDocumentHistory(userId int, docId string) (*[]models.DocumentHistory, error) {
	sql := `
	SELECT 
	    dh.id as id,
	    dh.document_id as document_id,
	    dh.action as action,
		dh.old_value as old_value,
		dh.new_value as new_value,
		dh.created_at as created_at,
		coalesce(dh.user_id, 0) as user_id,
		coalesce(u.name, 'Server') as user
	FROM document_history dh 
	LEFT JOIN documents d ON dh.document_id=d.id 
	LEFT JOIN users u ON dh.user_id=u.id
	WHERE document_id=$1
	ORDER BY created_at ASC;
	`

	data := &[]models.DocumentHistory{}
	err := s.db.Select(data, sql, docId)
	return data, s.parseError(err, "get document history")
}

func (s *DocumentStore) AddVisited(userId int, documentId string) error {
	query := s.sq.Insert("document_view_history").Columns("user_id", "document_id").Values(userId, documentId)
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("parse sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "add document_view_history")
}

func (s *DocumentStore) GetDocumentsInTrashbin(deletedAt time.Time) ([]string, error) {
	query := s.sq.Select("id").
		From("documents").
		Where("deleted_at < ?", deletedAt)
	sql, args, err := query.ToSql()
	ids := make([]string, 0)
	if err != nil {
		return ids, fmt.Errorf("sql: %v", err)
	}

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return ids, s.parseError(err, "get deleted documents")
	}

	for rows.Next() {
		var id = ""
		err = rows.Scan(&id)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (s *DocumentStore) BulkUpdateDocuments(userId int, docs []string, lang models.Lang, date time.Time) error {
	oldDocs, err := s.GetDocumentsById(userId, docs)
	if err != nil {
		return err
	}

	query := s.sq.Update("documents")
	if lang.String() != "" {
		query = query.Set("lang", lang)
	}
	if !date.IsZero() {
		query = query.Set("date", date)
	}

	query = query.Where("user_id = ?", userId).Where(squirrel.Eq{"id": docs})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	if err != nil {
		return getDatabaseError(err, s, "bulk update document lang")
	}

	updatedDocs, err := s.GetDocumentsById(userId, docs)
	if err != nil {
		return err
	}

	diffs := make([]models.DocumentHistory, 0)
	for i, v := range *oldDocs {
		newDoc := (*updatedDocs)[i]
		diff, err := v.Diff(&newDoc, userId)
		if err != nil {
			return fmt.Errorf("get diff for doc: %s: %v", v.Id, err)
		} else {
			diffs = append(diffs, diff...)
		}
	}
	err = addDocumentHistoryAction(s.db, s.sq, diffs, userId)
	if err != nil {
		return getDatabaseError(err, s, "insert document history")
	}

	err = s.setUpdatedAt(userId, docs, time.Now())
	return getDatabaseError(err, s, "update updated_at")
}

func (s *DocumentStore) setUpdatedAt(userId int, docs []string, updatedAt time.Time) error {
	query := s.sq.Update("documents").Set("updated_at", updatedAt).Where(squirrel.Eq{"id": docs})
	if userId != 0 {
		query = query.Where("user_id = ?", userId)
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	return getDatabaseError(err, s, "set updated_at")
}
