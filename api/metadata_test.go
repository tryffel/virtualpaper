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

package api

import (
	"testing"
)

func Test_metadataValueRequest_validate(t *testing.T) {
	type fields struct {
		Value          string
		Comment        string
		MatchDocuments bool
		MatchType      string
		MatchFilter    string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "simple case, no matching",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: false,
				MatchType:      "",
				MatchFilter:    "",
			},
			wantErr: false,
		},
		{
			name: "match, no filter defined",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: true,
				MatchType:      "",
				MatchFilter:    "",
			},
			wantErr: true,
		},
		{
			name: "match exact",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: true,
				MatchType:      "exact",
				MatchFilter:    "test",
			},
			wantErr: false,
		},
		{
			name: "valid regex",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: true,
				MatchType:      "regex",
				MatchFilter:    "(test)",
			},
			wantErr: false,
		},
		{
			name: "invalid regex",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: true,
				MatchType:      "regex",
				MatchFilter:    "(test)))",
			},
			wantErr: true,
		},
		{
			name: "invalid rule type",
			fields: fields{
				Value:          "test",
				Comment:        "",
				MatchDocuments: true,
				MatchType:      "test",
				MatchFilter:    "test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metadataValueRequest{
				Value:          tt.fields.Value,
				Comment:        tt.fields.Comment,
				MatchDocuments: tt.fields.MatchDocuments,
				MatchType:      tt.fields.MatchType,
				MatchFilter:    tt.fields.MatchFilter,
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
