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
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
	"testing"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func TestDocumentStore_GetDocument(t *testing.T) {

	db, mock, err := NewMockDatabase(sqlmock.QueryMatcherEqual)
	if err != nil {
		t.Fatal(err.Error())
	}

	doc := &models.Document{
		Timestamp: models.Timestamp{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Id:          "test-doc-id",
		UserId:      10,
		Name:        "test document",
		Description: "document description",
		Content:     "document content",
		Filename:    "file.pdf",
		Hash:        "1234",
		Mimetype:    "application/pdf",
		Size:        1024,
		Date:        time.Now().Add(-time.Hour * 24),
		Metadata:    nil,
		Tags:        nil,
	}

	mock.ExpectQuery("SELECT *\nFROM documents\nWHERE id = $1").
		WithArgs(doc.Id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "name", "description", "content", "filename", "hash",
			"mimetype", "size", "date", "created_at", "updated_at"}).
			AddRow(doc.Id, doc.UserId, doc.Name, doc.Description, doc.Content, doc.Filename, doc.Hash, doc.Mimetype,
				doc.Size, doc.Date, doc.CreatedAt, doc.UpdatedAt))

	gotDoc, err := db.DocumentStore.GetDocument(db, doc.Id)

	if err != nil {
		t.Error(err)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("invalid query: %v", err)
	}

	if !reflect.DeepEqual(doc, gotDoc) {
		t.Errorf("GetDocument() got = %v, want %v", gotDoc, doc)
	}
}
