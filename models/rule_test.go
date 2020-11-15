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

import "testing"

func TestRule_Validate(t *testing.T) {
	type fields struct {
		Type   RuleType
		Filter string
		Action RuleActionConfig
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		validate func(rule *Rule)
	}{
		{
			name: "invalid regex",
			fields: fields{
				Type:   "regex",
				Filter: "\\d((",
				Action: RuleActionConfig{},
			},
			wantErr:  true,
			validate: func(*Rule) {},
		},
		{
			name: "no action set",
			fields: fields{
				Type:   "regex",
				Filter: "\\d",
				Action: RuleActionConfig{},
			},
			wantErr:  true,
			validate: func(*Rule) {},
		},
		{
			name: "set metadata, invalid",
			fields: fields{
				Type:   "regex",
				Filter: "\\d",
				Action: RuleActionConfig{
					MetadataKeyId:   0,
					MetadataValueId: 10,
				},
			},
			wantErr:  true,
			validate: func(*Rule) {},
		},
		{
			name: "set metadata",
			fields: fields{
				Type:   "exact",
				Filter: "date",
				Action: RuleActionConfig{
					MetadataKeyId:   5,
					MetadataValueId: 10,
				},
			},
			wantErr: false,
			validate: func(r *Rule) {
				if !r.Action.Action.AddMetadata() {
					t.Error("metadata action not set")
				}
			},
		},
		{
			name: "add tag",
			fields: fields{
				Type:   "exact",
				Filter: "date",
				Action: RuleActionConfig{
					Tag: 10,
				},
			},
			wantErr: false,
			validate: func(r *Rule) {
				if !r.Action.Action.Tag() {
					t.Error("tag action not set")
				}

			},
		},
		{
			name: "set date",
			fields: fields{
				Type:   "exact",
				Filter: "date",
				Action: RuleActionConfig{
					DateFmt: "2006-01-02",
				},
			},
			wantErr: false,
			validate: func(r *Rule) {
				if !r.Action.Action.Date() {
					t.Error("date action not set")
				}
			},
		},
		{
			name: "set date",
			fields: fields{
				Type:   "exact",
				Filter: "date",
				Action: RuleActionConfig{
					Description: "desc",
				},
			},
			wantErr: false,
			validate: func(r *Rule) {
				if !r.Action.Action.Description() {
					t.Error("description action not set")
				}

			},
		},
		{
			name: "multiple actions",
			fields: fields{
				Type:   "exact",
				Filter: "date",
				Action: RuleActionConfig{
					Description: "desc",
					DateFmt:     "2006-01-02",
					Tag:         10,
				},
			},
			wantErr: false,
			validate: func(r *Rule) {
				if !r.Action.Action.Description() {
					t.Error("description action not set")
				}
				if !r.Action.Action.Date() {
					t.Error("date action not set")
				}
				if !r.Action.Action.Tag() {
					t.Error("tag action not set")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				Type:   tt.fields.Type,
				Filter: tt.fields.Filter,
				Action: tt.fields.Action,
			}
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.validate(r)
		})
	}
}
