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
				{DocumentId: "id", Action: "remove metadata", OldValue: "author:Bach", NewValue: ""},
				{DocumentId: "id", Action: "add metadata", OldValue: "", NewValue: "project:compsci"},
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
