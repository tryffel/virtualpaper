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
	"strings"
)

type Document struct {
	Timestamp
	Id       int    `db:"id"`
	UserId   int    `db:"user_id"`
	Name     string `db:"name"`
	Content  string `db:"content"`
	Filename string `db:"filename"`
	Hash     string `db:"hash"`
	Mimetype string `db:"mimetype"`
	Size     int64  `db:"size"`
}

// IsImage returns true if document file is image.
func (d *Document) IsImage() bool {
	return strings.Contains(d.Mimetype, "image/")
}

// IsPdf returns true id document file is pdf.
func (d *Document) IsPdf() bool {
	return strings.ToLower(d.Mimetype) == "application/pdf"
}

// GetThumbnail returns thumbnail file name
func (d *Document) GetThumbnailName() string {
	if d.Hash != "" {
		return d.Hash + ".png"
	} else {
		return ""
	}
}

func (d *Document) FilterAttributes() []string {
	ts := d.Timestamp.FilterAttributes()
	doc := []string{"id", "name", "content", "filename", "hash", "mimetype", "size"}
	return append(doc, ts...)
}

func (d *Document) SortAttributes() []string {
	return d.FilterAttributes()
}
