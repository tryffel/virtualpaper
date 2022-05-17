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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/otiai10/gosseract"
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

	logrus.Infof("Extract content for file %s with OCR", id)
	logrus.Debugf("convert pdf to images")

	imageFile := path.Join(dir, "preview.png")
	err = generatePicture(inputImage, imageFile)
	if err != nil {
		return text, fmt.Errorf("generate pictures from pdf pages: %v", err)
	}

	client := gosseract.NewClient()
	defer client.Close()
	err = client.SetLanguage(config.C.Processing.OcrLanguages...)
	if err != nil {
		logrus.Errorf("set tesseract languages: %v. continue with default language.", err)
	}
	pages := &[]string{}

	walkFunc := func(fileName string, info os.FileInfo, err error) error {
		if info.Name() == id {
			// root fileName
			return nil
		}
		start := time.Now()
		logrus.Infof("OCR file %s", fileName)
		err = client.SetImage(fileName)
		if err != nil {
			return fmt.Errorf("set ocr image source: %v", err)
		}
		pageText, err := client.Text()
		if err != nil {
			return fmt.Errorf("tesseract extract text: %v", err)
		}
		took := time.Now().Sub(start)
		logrus.Infof("Extracted %s, took %.2f s, content length: %d", fileName, took.Seconds(), len(pageText))
		*pages = append(*pages, pageText)
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
	return gosseract.Version()
}
