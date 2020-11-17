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
	"testing"
	"tryffel.net/go/virtualpaper/config"
)

func TestDocumentPath(t *testing.T) {
	config.C = &config.Config{
		Processing: config.Processing{
			DocumentsDir: "/data/documents",
		},
	}

	type args struct {
		documentId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{documentId: "3f24f12f-7977-4bae-8a22-3a304397b979"},
			want: "/data/documents/3/f/24f12f-7977-4bae-8a22-3a304397b979",
		},
		{
			args: args{documentId: "3f"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DocumentPath(tt.args.documentId); got != tt.want {
				t.Errorf("DocumentPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPreviewPath(t *testing.T) {
	config.C = &config.Config{
		Processing: config.Processing{
			PreviewsDir: "/data/previews",
		},
	}

	type args struct {
		documentId string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			args: args{documentId: "3f24f12f-7977-4bae-8a22-3a304397b979"},
			want: "/data/previews/3/f/24f12f-7977-4bae-8a22-3a304397b979",
		},
		{
			args: args{documentId: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PreviewPath(tt.args.documentId); got != tt.want {
				t.Errorf("PreviewPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
