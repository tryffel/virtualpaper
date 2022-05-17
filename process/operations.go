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
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"tryffel.net/go/virtualpaper/config"
)

func generateThumbnail(rawFile string, previewFile string, page int, size int, mimetype string) error {
	if mimetype == "text/plain" {
		return generateThumbnailPlainText(rawFile, previewFile, size)
	}
	logrus.Debugf("run 'convert -thumbnail'")

	args := []string{
		"-thumbnail", fmt.Sprintf("x%d", size),
		"-background", "white",
		//"-alpha", "remove",
		"-colorspace", "RGB",
		rawFile + fmt.Sprintf("[%d]", page),
		previewFile,
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("generate thumbnail: '%s'", strings.Join(args, " "))
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(config.C.Processing.ImagickBin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		logrus.Debugf("run %v: %v", args, err)
		return fmt.Errorf("execute convert: %v", err)
	}

	stdErr := stderr.String()
	if stdErr != "" {
		logrus.Warningf("Imagemagick failed, stderr: %v", err)
		return err
	}
	return nil
}

func generateThumbnailPlainText(rawFile string, previewFile string, size int) error {
	logrus.Debugf("generate thumbnail for text file")

	inputFile, err := os.Open(rawFile)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(previewFile)
	if err != nil {
		return fmt.Errorf("create output file: %v", err)
	}
	defer outputFile.Close()

	// A4 sized preview
	height := size
	width := int(float64(size) * 0.707)
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	y := 20

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, color.White)
		}
	}
	face := basicfont.Face7x13

	// split text to
	splitTextLines := func(text string) []string {
		maxY := 46

		if len(text) < maxY {
			return []string{text}
		}

		textLeft := text
		lines := make([]string, 0, 2)

		for true {
			if len(textLeft) == 0 {
				break
			}
			if len(textLeft) < maxY {
				lines = append(lines, textLeft)
				break
			}

			// find last whitespace to split at

			for i := maxY; i > 0; i-- {
				if textLeft[i] == ' ' {
					line := textLeft[0:i]
					line = strings.Trim(line, " \n")
					textLeft = textLeft[i+1:]
					lines = append(lines, line)
					break
				}
			}
		}

		return lines
	}

	// print text to file. Return true if print successful.
	// When page is full, return false.
	addText := func(text string) bool {
		splits := splitTextLines(text)

		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(color.RGBA{0, 0, 0, 255}),
			Face: face,
		}

		for _, row := range splits {
			// page full
			if y > height-20 {
				return false
			}

			d.Dot = fixed.Point26_6{fixed.Int26_6(4 * 64), fixed.Int26_6(y * 64)}
			d.DrawString(row)
			y += 20
		}

		return true
	}

	maxRows := 24
	maxChars := 1100

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	text := ""
	totalRows := 0

	for scanner.Scan() {
		row := scanner.Text()
		if text != "" {
			text += " "
		}
		totalRows += 1
		text += row
		if len(text) > maxChars || totalRows > maxRows {
			break
		}

		written := addText(row)
		if !written {
			// page full
			break
		}

	}

	err = png.Encode(outputFile, img)
	if err != nil {
		return fmt.Errorf("flush output buffer: %v", err)
	}
	return nil
}
