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
	"strings"
)

var mimeTypeToFileExtension map[string][]string

// constructed at build time, depending on installed binaries
// one mime type can refer to multiple extensions, e.g. text/plain -> txt & md

var fileExtensionToMimeType map[string]string

// constructed at build time, depending on installed binaries

var fileExtensionToName map[string]string

var pandocMimesSupported map[string]bool

func buildEmptyMimedataMapping() {
	mimeTypeToFileExtension = map[string][]string{}
	fileExtensionToMimeType = map[string]string{}
	fileExtensionToName = map[string]string{}
	pandocMimesSupported = map[string]bool{}
}

func buildMimeDataMapping() {
	buildEmptyMimedataMapping()

	// build supported mime data mapping
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

	if GetPandocInstalled() {
		csv := file{"text/csv", "csv", "Csv"}
		//json := file{"application/json", "json", "Json"}
		markdown := file{"text/plain", "md", "Markdown"}
		text := file{"text/plain", "txt", "Plain text"}
		//rst := file{""}
		docx := file{"application/vnd.openxmlformats-officedocument.wordprocessingml.document", "docx", "Word document"}
		msdoc := file{"application/msword", "doc", "Word document"}
		odt := file{"application/vnd.oasis.opendocument.text", "odt", "OpenDocument text document"}
		html := file{"text/html", "html", "Html"}
		epub := file{"application/epub+zip", "epub", "Epub (electronic publication book)"}

		supportedTypes = append(supportedTypes, csv, markdown, text, docx, msdoc, odt, html, epub)

		pandocMimesSupported = map[string]bool{
			csv.Mimetype:      true,
			markdown.Mimetype: true,
			text.Mimetype:     true,
			docx.Mimetype:     true,
			msdoc.Mimetype:    true,
			odt.Mimetype:      true,
			html.Mimetype:     true,
			epub.Mimetype:     true,
		}
	}

	for _, t := range supportedTypes {
		mimeTypeToFileExtension[t.Mimetype] = append(mimeTypeToFileExtension[t.Mimetype], t.Extension)
		fileExtensionToMimeType[t.Extension] = t.Mimetype
		fileExtensionToName[t.Extension] = t.Name
	}
}

// MimeTypeIsSupported returns true if mime type or file ending is supported.
// Either one, or both, can be filled. If both are "", return false.
// If both are filled, return true if expected file ending matches argument.
// If either one is filled, return true if that is supported.
func MimeTypeIsSupported(mimetype, filename string) bool {
	if mimetype == "" && filename == "" {
		return false
	}

	mimetype = strings.ToLower(mimetype)
	fileEnding := fileEndingFromName(filename)

	if mimetype != "" && fileEnding != "" {
		for _, v := range mimeTypeToFileExtension[mimetype] {
			if v == fileEnding {
				return true
			}
		}
		return false
	}

	if mimetype != "" {
		arr := mimeTypeToFileExtension[mimetype]
		return len(arr) > 0
	}

	if fileEnding != "" {
		return fileExtensionToMimeType[fileEnding] != ""
	}

	return false
}

func fileEndingFromName(filename string) string {
	if strings.Contains(filename, ".") {
		splits := strings.Split(filename, ".")
		fileEnding := splits[len(splits)-1]
		fileEnding = strings.ToLower(fileEnding)
		return fileEnding
	}
	return ""

}
