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
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
)

var pandocFileEndingtoFormat = map[string]string{
	"csv": "csv",
	"md":  "markdown",
	//"txt": "plain",
	"docx": "docx",
	"odt":  "odt",
	"html": "html",
	"epub": "epub",
}

func testPandoc() error {

	if config.C.Processing.PandocBin == "" {
		return errors.New("no pandoc binary set")
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(config.C.Processing.PandocBin, "-v")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("run pdftotext: %v", err)
	}

	stdErr := stderr.String()
	if stdErr != "" {
		if strings.Contains(stdErr, "pandoc version") {
			return nil
		} else {
			return fmt.Errorf("run pandoc: %s", stderr)
		}
	}

	result := stdout.String()
	if strings.Contains(result, "pandoc") {
		return nil
	} else {
		return fmt.Errorf("unknown pandoc version: %v", result)
	}
}

func GetPandocInstalled() bool {
	err := testPandoc()
	return err == nil
}

func isPandocMimetype(mimeType string) bool {
	return true
	return pandocMimesSupported[mimeType]
}

func getPandocText(mimetype, filename string, file *os.File) (string, error) {
	// todo: handle mime type as well

	fileEnding := fileEndingFromName(filename)

	if mimetype == "text/plain" {
		return readPlainTextFile(file)
	}

	format := pandocFileEndingtoFormat[fileEnding]

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(config.C.Processing.PandocBin, "-f", format, file.Name(), "-t", "plain")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	text := ""

	StdErr := stderr.String()
	if StdErr != "" {
		return "", fmt.Errorf("pandoc stderr: %v", StdErr)
	}

	err := cmd.Run()
	if err != nil {
		return text, fmt.Errorf("run pandoc: %v", err)
	}

	text = stdout.String()
	if text == "" {
		logrus.Warning("got empty text from pandoc")
	}

	return text, err
}

func readPlainTextFile(file *os.File) (string, error) {
	byteText, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return "", fmt.Errorf("read text file: %v", err)
	}

	return string(byteText), nil
}
