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
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func TestDocument_Diff(t *testing.T) {
	type fields struct {
		Id          string
		Name        string
		Description string
		Content     string
		Date        time.Time
		Metadata    []Metadata
	}
	type args struct {
		newDocument *Document
		userId      int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []DocumentHistory
		wantErr bool
	}{
		{
			name: "invalid id",
			fields: fields{
				Id: "abcd",
			},
			args:    args{newDocument: &Document{Id: "a"}},
			wantErr: true,
			want:    []DocumentHistory{},
		},
		{
			name: "change name",
			fields: fields{
				Id:   "id",
				Name: "testing",
			},
			args:    args{newDocument: &Document{Id: "id", Name: "testing2"}},
			wantErr: false,
			want:    []DocumentHistory{{DocumentId: "id", Action: "rename", OldValue: "testing", NewValue: "testing2"}},
		},
		{
			name: "change name and description",
			fields: fields{
				Id:          "id",
				Name:        "testing",
				Description: "empty",
			},
			args:    args{newDocument: &Document{Id: "id", Name: "testing2", Description: "description"}},
			wantErr: false,
			want: []DocumentHistory{
				{DocumentId: "id", Action: "rename", OldValue: "testing", NewValue: "testing2"},
				{DocumentId: "id", Action: "description", OldValue: "empty", NewValue: "description"},
			},
		},
		{
			name: "modify metadata",
			fields: fields{
				Id:          "id",
				Name:        "testing",
				Description: "empty",
				Metadata: []Metadata{
					{Key: "author", Value: "Bach"},
				},
			},
			args:    args{newDocument: &Document{Id: "id", Name: "testing2", Description: "description", Metadata: []Metadata{{Key: "project", Value: "compsci"}}}},
			wantErr: false,
			want: []DocumentHistory{
				{DocumentId: "id", Action: "rename", OldValue: "testing", NewValue: "testing2"},
				{DocumentId: "id", Action: "description", OldValue: "empty", NewValue: "description"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Document{
				Id:          tt.fields.Id,
				Name:        tt.fields.Name,
				Description: tt.fields.Description,
				Content:     tt.fields.Content,
				Date:        tt.fields.Date,
				Metadata:    tt.fields.Metadata,
			}
			got, err := d.Diff(tt.args.newDocument, tt.args.userId)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Diff() error = %v, wantErr %v", err, tt.wantErr)

			} else if err == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Document.Diff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDocument_Metadata_Diff(t *testing.T) {
	docId := "1234"
	userId := 10
	type fields struct {
		Metadata MetadataArray
	}
	type args struct {
		metadata MetadataArray
		userId   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []DocumentHistory
		wantErr bool
	}{
		{
			name: "modify metadata",
			fields: fields{
				Metadata: []Metadata{
					{KeyId: 2, ValueId: 4},
				},
			},
			args:    args{metadata: []Metadata{{KeyId: 10, ValueId: 15}}},
			wantErr: false,
			want: []DocumentHistory{
				{DocumentId: docId, UserId: userId, Action: "remove metadata", OldValue: `{"key_id":2,"value_id":4}`, NewValue: ""},
				{DocumentId: docId, UserId: userId, Action: "add metadata", OldValue: "", NewValue: `{"key_id":10,"value_id":15}`},
			},
		},
		{
			name: "modify partial metadata",
			fields: fields{
				Metadata: []Metadata{
					{KeyId: 2, ValueId: 4},
					{KeyId: 10, ValueId: 20},
					{KeyId: 11, ValueId: 21},
				},
			},
			args: args{metadata: []Metadata{
				{KeyId: 2, ValueId: 4},
				{KeyId: 10, ValueId: 22},
			}},
			wantErr: false,
			want: []DocumentHistory{
				{DocumentId: docId, UserId: userId, Action: "remove metadata", OldValue: `{"key_id":10,"value_id":20}`, NewValue: ""},
				{DocumentId: docId, UserId: userId, Action: "remove metadata", OldValue: `{"key_id":11,"value_id":21}`, NewValue: ""},
				{DocumentId: docId, UserId: userId, Action: "add metadata", OldValue: "", NewValue: `{"key_id":10,"value_id":22}`},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MetadataDiff(docId, userId, &tt.fields.Metadata, &tt.args.metadata)
			assert.Equal(t, got, tt.want)
		})
	}
}
