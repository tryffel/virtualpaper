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

package process

import (
	"testing"
)

func TestMimeTypeIsSupported(t *testing.T) {

	buildEmptyMimedataMapping()

	// add some standard mime type mapping.
	// we cannot use buildMimeDataMapping(), since it might produce different mappings in different setups.
	type file struct {
		Mimetype  string
		Extension string
		Name      string
	}
	supportedTypes := make([]file, 0, 4)

	pdf := file{"application/pdf", "pdf", "Pdf"}
	png := file{"image/png", "png", "Image"}
	jpg := file{"image/jpg", "jpg", "Image"}
	jpeg := file{"image/jpeg", "jpeg", "Image"}

	supportedTypes = append(supportedTypes, pdf, png, jpg, jpeg)

	for _, t := range supportedTypes {
		if mimeTypeToFileExtension[t.Mimetype] == nil {
			mimeTypeToFileExtension[t.Mimetype] = make([]string, 0, 1)
		}
		mimeTypeToFileExtension[t.Mimetype] = append(mimeTypeToFileExtension[t.Mimetype], t.Extension)
		fileExtensionToMimeType[t.Extension] = t.Mimetype
		fileExtensionToName[t.Extension] = t.Name
	}

	tests := []struct {
		name     string
		mime     string
		filename string
		want     bool
	}{
		{
			name:     "mime and file name ok",
			mime:     "application/pdf",
			filename: "test.pdf",
			want:     true,
		},
		{
			name: "mime ok",
			mime: "application/pdf",
			want: true,
		},
		{
			name: "mime uppercased",
			mime: "application/PDF",
			want: true,
		},
		{
			name:     "name uppercased",
			filename: "TEST.PDF",
			want:     true,
		},
		{
			name: "invalid mime",
			mime: "application/pdfa",
			want: false,
		},
		{
			name:     "file name ok",
			filename: "test.pdf",
			want:     true,
		},
		{
			name:     "invalid filename",
			filename: "test.pdfa",
			want:     false,
		},
		{
			name:     "invalid filename",
			filename: "test",
			want:     false,
		},
		{
			name: "empty query",
			want: false,
		},
		{
			name:     "valid mime, invalid name",
			mime:     "application/pdf",
			filename: "test.pdfa",
			want:     false,
		},
		{
			name:     "valid mime, no file ending",
			mime:     "application/pdf",
			filename: "test",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MimeTypeIsSupported(tt.mime, tt.filename); got != tt.want {
				t.Errorf("MimeTypeIsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
