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

type DocumentStore struct {
	db *sqlx.DB
}

func (s *DocumentStore) GetDocuments(userId int, limit int) (*[]models.Document, error) {
	sql := `
SELECT id, name, filename, created_at, updated_at 
FROM documents
WHERE user_id = $1
LIMIT $2;
`

	dest := &[]models.Document{}
	err := s.db.Select(dest, sql, userId, limit)

	return dest, err
}
