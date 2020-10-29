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
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Document struct {
	Timestamp
	Id          int       `db:"id"`
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

func (d *Document) GetSize() string {
	if d.Size < 1024 {
		return strconv.Itoa(int(d.Size))
	}
	size := float64(d.Size)
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f KiB", size)
	}
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f MiB", size)
	}
	size /= 1024
	if size < 1024 {
		return fmt.Sprintf("%.2f GiB", size)
	}
	return fmt.Sprintf("%f B", size)

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

type Metadata struct {
	KeyId   int    `db:"key_id" json:"key_id"`
	Key     string `db:"key" json:"key"`
	ValueId int    `db:"value_id" json:"value_id"`
	Value   string `db:"value" json:"value"`
}

type MetadataKey struct {
	Id        int       `db:"id" json:"id"`
	UserId    string    `db:"user_id" json:"-"`
	Key       string    `db:"key" json:"key"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Comment   string    `db:"comment" json:"comment"`
}

type MetadataValue struct {
	Id        int       `db:"id" json:"id"`
	UserId    string    `db:"user_id" json:"-"`
	Key       string    `db:"key" json:"key"`
	KeyId     int       `db:"key_id" json:"-"`
	Value     string    `db:"value" json:"value"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Comment   string    `db:"comment" json:"comment"`
}
