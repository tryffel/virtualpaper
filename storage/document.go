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
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
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
func (s *DocumentStore) GetDocuments(exec SqlExecer, userId int, paging Paging, sort SortKey, limitContent bool, showTrash bool, showSharesDocs bool) (*[]models.Document, int, error) {
	sort.SetDefaults("date", false)
	var contentSelect string
	if limitContent {
		contentSelect = "LEFT(content, 500) as content"
	} else {
		contentSelect = "content"
	}

	query := s.sq.Select("id, name, filename, documents.created_at as created_at, documents.updated_at as updated_at, hash, mimetype, size, date, description, lang, deleted_at, count(shares.user_id) as shares, favorite", contentSelect).
		From("documents").
		LeftJoin("user_shared_documents shares on documents.id = shares.document_id")

	var trashQuery squirrel.Sqlizer
	var ownerQuery squirrel.Sqlizer
	isOwnerQuery := squirrel.Eq{"documents.user_id": userId}
	if showTrash {
		trashQuery = squirrel.NotEq{"deleted_at": nil}
	} else {
		trashQuery = squirrel.Eq{"deleted_at": nil}
	}

	if showSharesDocs {
		sharedQuery := squirrel.Expr(`documents.id IN ( SELECT share.document_id 
FROM user_shared_documents share
WHERE share.user_id = ? AND (share.permission -> 'read')::boolean = true)`, userId)

		ownerQuery = squirrel.Or{
			isOwnerQuery,
			sharedQuery,
		}
	} else {
		ownerQuery = isOwnerQuery
	}

	query = query.Where(squirrel.And{
		trashQuery,
		ownerQuery,
	})
	query = query.GroupBy("documents.id")
	query = query.OrderBy(fmt.Sprintf("%s %s", sort.QueryKey(), sort.SortOrder()))
	query = query.Offset(uint64(paging.Offset)).Limit(uint64(paging.Limit))

	dest := &[]models.Document{}

	err := exec.SelectSq(dest, query)
	if limitContent && len(*dest) > 0 {
		for i, _ := range *dest {
			if len((*dest)[i].Content) > 499 {
				(*dest)[i].Content += "..."
			}
		}
	}

	query = s.sq.Select("count(id)").From("documents").Where("user_id = ?", userId)
	if showTrash {
		query = query.Where(squirrel.NotEq{"deleted_at": nil})
	} else {
		query = query.Where(squirrel.Eq{"deleted_at": nil})
	}

	var count int
	err = exec.GetSq(&count, query)
	err = s.parseError(err, "get documents")
	return dest, count, err
}

// GetDocument returns document by its id.
func (s *DocumentStore) GetDocument(execer SqlExecer, id string) (*models.Document, error) {
	sql := `
SELECT *
FROM documents
WHERE id = $1
`
	dest := &models.Document{}
	err := execer.Get(dest, sql, id)
	return dest, s.parseError(err, "get document")
}

func (s *DocumentStore) GetSharedUsers(exec SqlExecer, docId string) (*[]models.DocumentSharePermission, error) {
	query := s.sq.Select("users.id as user_id, users.name as user_name, share.document_id as document_id, share.permission as permissions").
		From("user_shared_documents share").
		LeftJoin("users on share.user_id = users.id").
		Where("document_id = ?", docId)

	dest := &[]models.DocumentSharePermission{}
	err := exec.SelectSq(dest, query)
	return dest, s.parseError(err, "get document shared users")
}

// GetDocument returns document by its id. If userId != 0, user must be owner of the document.
func (s *DocumentStore) GetDocumentsById(exec SqlExecer, userId int, id []string) (*[]models.Document, error) {

	query := s.sq.Select("*").From("documents").Where(squirrel.Eq{"id": id})
	if userId != 0 {
		query = query.Where("user_id = ?", userId)
	}
	query = query.OrderBy("id ASC")
	dest := &[]models.Document{}
	err := exec.SelectSq(dest, query)
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

func (s *DocumentStore) GetPermissions(exec SqlExecer, documentId string, userId int) (owner bool, perm models.Permissions, err error) {
	type Result struct {
		Owner      bool               `db:"owner"`
		DocumentId string             `db:"document_id"`
		Permission models.Permissions `db:"permissions"`
	}

	sql := `
SELECT doc.id AS document_id, doc.owner AS owner, share.permission AS permissions
FROM (
         SELECT doc.user_id = $1 AS owner, doc.id
         FROM documents doc
         WHERE doc.id = $2
     ) AS doc
LEFT JOIN
    (
        SELECT share.document_id, share.user_id, share.permission
        FROM documents doc
			LEFT JOIN user_shared_documents share ON doc.id = share.document_id
        WHERE doc.id = $2 AND share.user_id = $1
    ) AS share ON doc.id = share.document_id;`

	result := &Result{}
	err = exec.Get(result, sql, userId, documentId)
	if err != nil {
		err = getDatabaseError(err, s, "get permissions")
		if errors.Is(err, errors.ErrRecordNotFound) {
			perm = models.Permissions{
				Read:   false,
				Write:  false,
				Delete: false,
			}
		}
		return
	}
	perm = result.Permission
	owner = result.Owner
	return
}

func (s *DocumentStore) UserOwnsDocuments(queries SqlExecer, userId int, documents []string) (bool, error) {
	query := s.sq.Select("count(distinct(id))").From("documents").Where("user_id = ?", userId).Where(
		squirrel.Eq{"id": documents})

	var documentCount int
	err := queries.GetSq(&documentCount, query)
	if err != nil {
		return false, s.parseError(err, "check user owns documents")
	}

	return documentCount == len(documents), s.parseError(err, "check user owns documents")
}

// GetByHash returns a document by its hash and user.
func (s *DocumentStore) GetByHash(userId int, hash string) (*models.Document, error) {
	query := s.sq.Select("*").From("documents").Where("hash = ?", hash).Where("user_id = ?", userId)
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}
	object := &models.Document{}
	err = s.db.Get(object, sql, args...)
	if err != nil {
		e := s.parseError(err, "get by hash")
		if errors.Is(e, errors.ErrRecordNotFound) {
			return object, nil
		}
		return object, e
	}
	return object, nil
}

func (s *DocumentStore) Create(exec SqlExecer, doc *models.Document) error {
	sql := `
INSERT INTO documents (id, user_id, name, content, filename, hash, mimetype, size, description, date, lang, favorite)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id;`

	doc.Init()

	rows, err := s.db.Query(sql, doc.Id, doc.UserId, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Description, doc.Date, doc.Lang, doc.Favorite)
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
	err = AddDocumentHistoryAction(exec, s.sq, []models.DocumentHistory{{DocumentId: doc.Id, Action: models.DocumentHistoryActionCreate, OldValue: "", NewValue: doc.Name}}, doc.UserId)
	return err
}

func AddDocumentHistoryAction(exec SqlExecer, queryBuilder squirrel.StatementBuilderType, items []models.DocumentHistory, userId int) error {
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

	_, err := exec.ExecSq(query)
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
func (s *DocumentStore) GetContent(id string) (*string, error) {
	sql := `
SELECT content
FROM documents
WHERE id=$1
`
	content := ""
	var err error
	err = s.db.Get(&content, sql, id)
	return &content, s.parseError(err, "get content")
}

// Update sets complete document record, not just changed attributes. Thus document must be read before updating.
// Metadata history is not saved.
func (s *DocumentStore) Update(exec SqlExecer, userId int, doc *models.Document) error {
	oldDoc, err := s.GetDocument(exec, doc.Id)
	if err != nil {
		return s.parseError(err, "get document by id")
	}
	doc.UpdatedAt = time.Now()
	sql := `
UPDATE documents SET 
name=$2, content=$3, filename=$4, hash=$5, mimetype=$6, size=$7, date=$8,
updated_at=$9, description=$10, lang=$11, favorite=$12
WHERE id=$1
`

	_, err = exec.Exec(sql, doc.Id, doc.Name, doc.Content, doc.Filename, doc.Hash, doc.Mimetype, doc.Size,
		doc.Date, doc.UpdatedAt, doc.Description, doc.Lang, doc.Favorite)
	if err != nil {
		return s.parseError(err, "update")
	}

	diff, err := oldDoc.Diff(doc, userId)
	if err != nil {
		return fmt.Errorf("get diff for document: %v", err)
	}

	err = AddDocumentHistoryAction(exec, s.sq, diff, userId)
	logrus.Infof("User %d edited document %s with %d actions", userId, doc.Id, len(diff))
	return err
}

func (s *DocumentStore) SetModifiedAt(exec SqlExecer, docIds []string, modifiedAt time.Time) error {
	query := s.sq.Update("documents").Set("updated_at", modifiedAt).Where(squirrel.Eq{"id": docIds})
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = exec.Exec(sql, args...)
	return s.parseError(err, "update modified_at")
}

func (s *DocumentStore) MarkDocumentDeleted(exec SqlExecer, userId int, docId string) error {
	query := s.sq.Update("documents").Set("deleted_at", time.Now()).Where("id=?", docId)
	if userId != 0 {
		query = query.Where("user_id = ?", userId)
	}

	_, err := exec.ExecSq(query)
	if err != nil {
		return s.parseError(err, "mark document deleted")
	}

	err = AddDocumentHistoryAction(exec, s.sq, []models.DocumentHistory{{
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

func (s *DocumentStore) MarkDocumentNonDeleted(exec SqlExecer, userId int, docId string) error {
	query := s.sq.Update("documents").Set("deleted_at", nil).Where("id=?", docId)
	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)

	if err != nil {
		return s.parseError(err, "mark document deleted")
	}
	err = AddDocumentHistoryAction(exec, s.sq, []models.DocumentHistory{{
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
		rows.Close()
	}
	return ids, nil
}

func (s *DocumentStore) BulkUpdateDocuments(exec SqlExecer, userId int, docs []string, lang models.Lang, date time.Time) error {
	oldDocs, err := s.GetDocumentsById(exec, userId, docs)
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

	updatedDocs, err := s.GetDocumentsById(exec, userId, docs)
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
	err = AddDocumentHistoryAction(exec, s.sq, diffs, userId)
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

func (s *DocumentStore) UpdateSharing(exec SqlExecer, docId string, sharing *[]models.UpdateUserSharing) error {
	doc, err := s.GetDocument(exec, docId)
	if err != nil {
		return err
	}

	for _, v := range *sharing {
		if doc.UserId == v.UserId {
			userErr := errors.ErrInvalid
			userErr.ErrMsg = "cannot share with self"
			return userErr
		}
	}

	query := s.sq.Delete("user_shared_documents").Where("document_id = ?", docId)
	_, err = exec.ExecSq(query)
	if err != nil {
		return fmt.Errorf("delete existing record: %v", err)
	}

	if len(*sharing) > 0 {
		insertQuery := s.sq.Insert("user_shared_documents").Columns("user_id", "document_id", "permission")

		for _, v := range *sharing {
			insertQuery = insertQuery.Values(v.UserId, docId, v.Permissions)
		}

		_, err = exec.ExecSq(insertQuery)
		dbErr := getDatabaseError(err, s, "update user_shared_documents")
		if dbErr == nil {
			return nil
		}

		if errors.Is(dbErr, errors.ErrRecordNotFound) {
			noUserErr := errors.ErrRecordNotFound
			noUserErr.ErrMsg = "user not found"
			return noUserErr
		}
	}
	return nil
}
