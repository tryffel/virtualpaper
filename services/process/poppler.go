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
	"bytes"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
)

// test pdftotext command exists
func testPdfToText() error {
	if config.C.Processing.PdfToTextBin == "" {
		return errors.New("no pdftotext binary set")
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(config.C.Processing.PdfToTextBin, "-v")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run pdftotext: %v", err)
	}

	stdErr := stderr.String()
	if stdErr != "" {
		if strings.Contains(stdErr, "pdftotext version") {
			return nil
		} else {
			return fmt.Errorf("run pdftotext: %s", stderr)
		}
	}

	result := stdout.String()
	if strings.Contains(result, "pdftotext version") {
		return nil
	} else {
		return fmt.Errorf("unknown pdftotext version: %v", result)
	}
}

func GetPdfToTextIsInstalled() bool {
	err := testPdfToText()
	return err == nil
}

// try to convert pdf to text directly without ocr. If pdf does not contain any text, return err
// 'empty'. Hash is used for temporary file
func getPdfToText(file *os.File, id string) (string, error) {
	textFile := storage.TempFilePath(id) + ",txt"
	defer removeTempData(textFile)

	text := ""

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := exec.Command(config.C.Processing.PdfToTextBin, file.Name(), textFile)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return text, fmt.Errorf("run pdftotext: %v", err)
	}

	result := stdout.String()
	if result != "" {
		logrus.Warningf("got non-zero result from pdftotext: %v", result)
	}

	StdErr := stderr.String()
	if StdErr != "" {
		return "", fmt.Errorf("pdftotext stderr: %v", err)
	}

	byteText, err := ioutil.ReadFile(textFile)
	if err != nil {
		return text, fmt.Errorf("read text file: %v", err)
	}

	text = string(byteText)
	if len(text) < 5 {
		return text, errors.New("empty")
	}

	return string(byteText), nil
}

// try to remote temp file. If file does not exist, do nothing. Else in case of errors log error.
func removeTempData(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
		} else {
			logrus.Errorf("remove temporary file: %v", err)
		}
	}
}
