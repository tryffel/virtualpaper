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
