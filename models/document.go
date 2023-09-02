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

package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
)

// Document represents single file and data related to it.
type Document struct {
	Timestamp
	Id          string    `db:"id"`
	UserId      int       `db:"user_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	Content     string    `db:"content"`
	Filename    string    `db:"filename"`
	Hash        string    `db:"hash"`
	Mimetype    string    `db:"mimetype"`
	Size        int64     `db:"size"`
	Date        time.Time `db:"date"`
	Metadata    []Metadata
	Tags        []Tag
	Lang        Lang `db:"lang"`

	DeletedAt sql.NullTime `db:"deleted_at"`
}

// Init initializes new document. It ensures document has valid uuid assigned to it.
func (d *Document) Init() {
	if d.Id == "" {
		var err error
		d.Id, err = uuid.GenerateUUID()
		if err == nil {
			return
		}
		logrus.Warningf("failed to generate uuid: %v, retrying", err)
		d.Id, err = uuid.GenerateUUID()
		if err == nil {
			return
		}
		logrus.Errorf("generate uuid: %v. Assign random string as id ", err)
		d.Id, err = config.RandomString(16)
		if err != nil {
			// this is okay, if the id exists in db then the document will be discarded.
			logrus.Errorf("generate document id by random string %v", err)
		}
	}
	if d.Date.IsZero() {
		d.Date = time.Now()
	}
}

// IsImage returns true if document file is image.
func (d *Document) IsImage() bool {
	return strings.Contains(d.Mimetype, "image/")
}

// IsPdf returns true id document file is pdf.
func (d *Document) IsPdf() bool {
	return strings.ToLower(d.Mimetype) == "application/pdf"
}

// GetType returns either 'pdf' or 'image' depending on type of content.
func (d *Document) GetType() string {
	if d.IsPdf() {
		return "pdf"
	} else if d.IsImage() {
		return "image"
	} else {
		return d.Mimetype
	}
}

// GetSize returns human-formatted size
func (d *Document) GetSize() string {
	return GetPrettySize(d.Size)
}

func (d *Document) FilterAttributes() []string {
	ts := d.Timestamp.FilterAttributes()
	doc := []string{"id", "name", "content", "description", "filename", "hash", "mimetype", "size", "date", "deleted_at", "created_at", "updated_at"}
	return append(doc, ts...)
}

func (d *Document) SortAttributes() []string {
	return d.FilterAttributes()
}

func (d *Document) SortNoCase() []string {
	return []string{"name", "content", "description", "filename", "hash", "mimetype"}
}

// HasMetadataKey returns true if document has given metadata key.
func (d *Document) HasMetadataKey(keyId int) bool {
	for _, v := range d.Metadata {
		if v.KeyId == keyId {
			return true
		}
	}
	return false
}

// HasMetadataKeyValue returns true if document has given metadata key and value.
func (d *Document) HasMetadataKeyValue(keyId, valueId int) bool {
	for _, v := range d.Metadata {
		if v.KeyId == keyId && v.ValueId == valueId {
			return true
		}
	}
	return false
}

// Metadata is metadata key-value pair assigned to document
type Metadata struct {
	KeyId   int    `db:"key_id" json:"key_id"`
	Key     string `db:"key" json:"key"`
	ValueId int    `db:"value_id" json:"value_id"`
	Value   string `db:"value" json:"value"`
}

type MetadataKey struct {
	Id        int       `db:"id" json:"id"`
	UserId    int       `db:"user_id" json:"-"`
	Key       string    `db:"key" json:"key"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Comment   string    `db:"comment" json:"comment"`
}

func MetadataDiff(id string, userId int, original, updated *[]Metadata) []DocumentHistory {
	history := make([]DocumentHistory, 0)

	addHistoryItem := func(action, oldValue, newValue string) {
		history = append(history, DocumentHistory{
			DocumentId: id,
			Action:     action,
			OldValue:   oldValue,
			NewValue:   newValue,
			UserId:     userId,
		})
	}

	if len(*original) == 0 && len(*updated) == 0 {
		return history
	}

	oldMetadata := map[string]Metadata{}
	newMetadata := map[string]Metadata{}

	for _, v := range *original {
		oldMetadata[fmt.Sprintf("%d-%d", v.KeyId, v.ValueId)] = v
	}

	for _, v := range *updated {
		newMetadata[fmt.Sprintf("%d-%d", v.KeyId, v.ValueId)] = v
	}

	formatMetadata := func(m Metadata) string {
		data := DocumentMetadataHistoryEntry{
			KeyId:   m.KeyId,
			ValueId: m.ValueId,
		}
		bytes, err := json.Marshal(data)
		if err != nil {
			logrus.Errorf("metadatadiff, marshal metadata to json, key: %d, value: %d: %v", m.KeyId, m.ValueId, err)
			return "error"
		}
		return string(bytes)
	}

	for keyValue, oldVal := range oldMetadata {
		if _, found := newMetadata[keyValue]; !found {
			addHistoryItem(DocumentHistoryActionMetadataRemove, formatMetadata(oldVal), "")
		}
	}

	for keyValue, newVal := range newMetadata {
		if _, found := oldMetadata[keyValue]; !found {
			addHistoryItem(DocumentHistoryActionMetadataAdd, "", formatMetadata(newVal))
		}
	}
	return history
}

type MetadataKeyStatistics struct {
	NumDocuments      int `db:"documents_count" json:"documents_count"`
	NumMetadataValues int `db:"values_count" json:"metadata_values_count"`
}

type MetadataKeyAnnotated struct {
	MetadataKey
	MetadataKeyStatistics
}

func (m *MetadataKey) Update() {}

func (m *MetadataKey) FilterAttributes() []string {
	return []string{"id", "key", "created_at", "comment", "documents_count"}
}

func (m *MetadataKey) SortNoCase() []string {
	return []string{"key", "comment"}
}

func (m *MetadataKey) SortAttributes() []string {
	return m.FilterAttributes()
}

type MetadataValue struct {
	Id           int       `db:"id" json:"id"`
	UserId       int       `db:"user_id" json:"-"`
	Key          string    `db:"key" json:"key"`
	KeyId        int       `db:"key_id" json:"-"`
	Value        string    `db:"value" json:"value"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	Comment      string    `db:"comment" json:"comment"`
	NumDocuments int       `db:"documents_count" json:"documents_count"`

	// MatchDocuments instructs to try to automatically match MetadataValue inside documents
	MatchDocuments bool             `db:"match_documents" json:"match_documents"`
	MatchType      MetadataRuleType `db:"match_type" json:"match_type"`
	MatchFilter    string           `db:"match_filter" json:"match_filter"`
}

func (m *MetadataValue) Update() {}

func (m *MetadataValue) FilterAttributes() []string {
	return []string{"id", "key", "value", "created_at", "comment", "documents_count",
		"match_documents", "match_type", "match_filter"}
}

func (m *MetadataValue) SortAttributes() []string {
	return m.FilterAttributes()
}

func (m *MetadataValue) SortNoCase() []string {
	return []string{"key", "value", "comment", "match_filter"}
}

type DocumentHistory struct {
	Id         int       `db:"id" json:"id"`
	DocumentId string    `db:"document_id" json:"document_id"`
	Action     string    `db:"action" json:"action"`
	OldValue   string    `db:"old_value" json:"old_value"`
	NewValue   string    `db:"new_value" json:"new_value"`
	UserId     int       `db:"user_id" json:"user_id"`
	User       string    `db:"user" json:"user"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

func (dh *DocumentHistory) Update() {}

func (dh *DocumentHistory) FilterAttributes() []string {
	return []string{"id", "document_id", "action"}
}

func (dh *DocumentHistory) SortAttributes() []string {
	return dh.FilterAttributes()
}

func (dh *DocumentHistory) SortNoCase() []string {
	return dh.FilterAttributes()
}

type DocumentMetadataHistoryEntry struct {
	KeyId   int `json:"key_id"`
	ValueId int `json:"value_id"`
}

type DocumentHistoryAction string

const (
	DocumentHistoryActionCreate         = "create"
	DocumentHistoryActionRename         = "rename"
	DocumentHistoryActionDescription    = "description"
	DocumentHistoryActionDate           = "date"
	DocumentHistoryActionLanguage       = "lang"
	DocumentHistoryActionContent        = "content"
	DocumentHistoryActionMetadataRemove = "remove metadata"
	DocumentHistoryActionMetadataAdd    = "add metadata"
	DocumentHistoryActionDelete         = "delete"
	DocumentHistoryActionRestore        = "restore"
)

// Diffs returns a list of DocumentHistory items from d -> newDocument.
// Metadata changes are not evaluated.
func (d *Document) Diff(newDocument *Document, userId int) ([]DocumentHistory, error) {
	history := make([]DocumentHistory, 0)
	d2 := newDocument

	addHistoryItem := func(action, oldValue, newValue string) {
		history = append(history, DocumentHistory{
			DocumentId: d2.Id,
			Action:     action,
			OldValue:   oldValue,
			NewValue:   newValue,
			UserId:     userId,
		})
	}

	if d.Id != newDocument.Id {
		return nil, errors.New("document id does not match")
	}

	if d.Name != d2.Name {
		addHistoryItem(DocumentHistoryActionRename, d.Name, d2.Name)
	}
	if d.Description != d2.Description {
		addHistoryItem(DocumentHistoryActionDescription, d.Description, d2.Description)
	}

	if MidnightForDate(d.Date) != MidnightForDate(d2.Date) {
		addHistoryItem(DocumentHistoryActionDate, strconv.Itoa(int(d.Date.Unix())), strconv.Itoa(int(d2.Date.Unix())))
	}

	if d.Content != d2.Content {
		addHistoryItem(DocumentHistoryActionContent, d.Content, d2.Content)
	}
	if d.Lang != d2.Lang {
		addHistoryItem(DocumentHistoryActionLanguage, d.Lang.String(), d2.Lang.String())
	}
	return history, nil
}

// LinkedDocument represents documents that are linked together
type LinkedDocument struct {
	DocumentId   string    `json:"id"`
	DocumentName string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
}
