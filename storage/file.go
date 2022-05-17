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
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"tryffel.net/go/virtualpaper/config"
)

// DocumentPath returns path for document by its id. Function
// splits documents to 2-level directories inside config.C.Processing.DocumentsDir.
// Id must be at least 3 characters long, else empty string is returned.
func DocumentPath(documentId string) string {
	if len(documentId) < 3 {
		return ""
	}

	dir0 := string(documentId[0])
	dir1 := string(documentId[1])
	rest := documentId[2:]
	out := path.Join(config.C.Processing.DocumentsDir, dir0, dir1, rest)
	return out
}

// CreateDocumentDir creates (if not yet existing) directory for document.
func CreateDocumentDir(documentId string) error {
	path := path.Dir(DocumentPath(documentId))
	err := os.MkdirAll(path, 0755|os.ModeSetgid|os.ModeSetuid)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}

// DocumentPath returns path for document preview by its id. Function
// splits previews to 2-level directories inside config.C.Processing.PreviewsDir.
// Id must be at least 3 characters long, else empty string is returned.
func PreviewPath(documentId string) string {
	if len(documentId) < 3 {
		return ""
	}

	dir0 := string(documentId[0])
	dir1 := string(documentId[1])
	rest := documentId[2:]
	out := path.Join(config.C.Processing.PreviewsDir, dir0, dir1, rest) + ".png"
	return out
}

// CreatePreviewDir creates (if not yet existing) directory for preview.
func CreatePreviewDir(documentId string) error {
	path := path.Dir(PreviewPath(documentId))
	err := os.MkdirAll(path, 0755|os.ModeSetgid|os.ModeSetuid)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	return nil
}

// TempFilePath returns filename in temporary directory for given id.
func TempFilePath(documentId string) string {
	return path.Join(config.C.Processing.TmpDir, documentId)
}

// MoveFile moves file from old location to new. It copies file, if necessary.
func MoveFile(from string, to string) error {
	err := os.Rename(from, to)
	if err == nil {
		return nil
	}

	if errors.Is(err, os.ErrExist) || errors.Is(err, os.ErrNotExist) {
		return err
	}

	e := err.Error()
	if strings.Contains(e, "invalid cross-device link") {
		oldFile, err := os.Open(from)
		if err != nil {
			return err
		}

		newFile, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}

		n, err := io.Copy(newFile, oldFile)
		if err != nil {
			return fmt.Errorf("copy data from old file: %v", err)
		}
		newFile.Close()

		stat, err := oldFile.Stat()
		if err != nil {
			return fmt.Errorf("get file stat: %v", err)
		}
		if stat.Size() != n {
			return fmt.Errorf("did not fully copy file: expect %d bytes, copied %d bytes", stat.Size(), n)
		}

		oldFile.Close()
		err = os.Remove(from)
		return err
	}
	return err
}
