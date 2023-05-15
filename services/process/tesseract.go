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
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
)

func runOcr(inputImage, id string) (string, error) {

	var err error
	var text string

	dir := storage.TempFilePath(id)
	err = os.Mkdir(dir, os.ModePerm|os.ModeDir)
	if err != nil {
		return text, fmt.Errorf("create tmp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	logrus.Infof("Extract content for file %s with OCR", id)
	logrus.Debugf("convert pdf to images")

	imageFile := path.Join(dir, "preview.png")
	err = generatePicture(inputImage, imageFile)
	if err != nil {
		return text, fmt.Errorf("generate pictures from pdf pages: %v", err)
	}

	pages := &[]string{}

	languageParam := strings.Join(config.C.Processing.OcrLanguages, "+")

	walkFunc := func(fileName string, info os.FileInfo, err error) error {
		if info.Name() == id {
			// root fileName
			return nil
		}
		start := time.Now()
		logrus.Infof("OCR file %s", fileName)

		outputFile := fileName + "-out"

		args := []string{
			fileName,
			outputFile,
			"-l",
			languageParam,
		}

		_, err = callTesseract(args...)
		if err != nil {
			logrus.Errorf("call tesseract: %s -  %v", args, err)
		}

		output, err := os.Open(outputFile + ".txt")
		if err != nil {
			logrus.Errorf("read output file %s: %v", outputFile, err)
		}

		pageText, err := io.ReadAll(output)
		err = output.Close()
		if err != nil {
			logrus.Errorf("close output file %s: %v", outputFile, err)

		}

		took := time.Now().Sub(start)
		logrus.Infof("Extracted %s, took %.2f s, content length: %d", fileName, took.Seconds(), len(pageText))
		*pages = append(*pages, string(pageText))
		return nil
	}

	err = filepath.Walk(dir, walkFunc)

	if err != nil {
		return text, fmt.Errorf("ocr file: %v", err)
	}

	content := ""
	for i, v := range *pages {
		if i > 0 {
			content += fmt.Sprintf("\n\n(Page %d)\n\n", i)
		}
		content += v
	}
	return content, err
}

func GetTesseractVersion() string {
	out, err := callTesseract("--version")
	if err != nil {
		logrus.Error(err)
	}

	splits := strings.Split(out, "\n")
	if len(splits) == 0 {
		return out
	}
	return splits[0]
}

func callTesseract(args ...string) (string, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	logrus.Debugf("call tesseract: %s, %v", config.C.Processing.TesseractBin, args)
	cmd := exec.Command(config.C.Processing.TesseractBin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	stdErr := stderr.String()
	if stdErr != "" {
		if strings.HasPrefix(stdErr, "Estimating resolution") {
			// skip
			logrus.Warningf("Tesseract warning, stderr: %v", stderr)
		} else {
			logrus.Warningf("Tesseract failed, stderr: %v", stderr)
			return stdErr, err
		}
	}
	if err != nil {
		logrus.Warningf("run %v: %v", args, err)
		return stdErr, fmt.Errorf("call tesseract: %v", err)
	}
	return stdout.String(), nil
}
