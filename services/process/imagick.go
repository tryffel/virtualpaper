/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
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
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"tryffel.net/go/virtualpaper/config"
	log "tryffel.net/go/virtualpaper/util/logger"
)

var imagickRe = `Version: ImageMagick\s(.+)`

func GetImagickVersion() string {
	regex := regexp.MustCompile(imagickRe)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cmd := exec.Command(config.C.Processing.ImagickBin, "-version")
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		logrus.Warningf("Imagemagick not found: %v", err)
		return ""
	}

	stdErr := stderr.String()
	if stdErr != "" {
		logrus.Warningf("Imagemagick not found, stderr: %v", err)
		return ""
	}

	ver := regex.FindString(stdout.String())
	logrus.Debugf("found imagemagick version %s", ver)
	return ver
}

func callImagick(args ...string) error {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	logrus.Debugf("call imagick: %s, %v", config.C.Processing.ImagickBin, args)
	cmd := exec.Command(config.C.Processing.ImagickBin, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	stdErr := stderr.String()
	if stdErr != "" {
		logrus.Warningf("Imagemagick failed, stderr: %v", err)
		return err
	}
	if err != nil {
		logrus.Warningf("run %v: %v", args, err)
		return fmt.Errorf("execute convert: %v", err)
	}
	return nil
}

func generateThumbnail(ctx context.Context, rawFile string, previewFile string, page int, size int, mimetype string) error {
	if mimetype == "text/plain" {
		return generateThumbnailPlainText(rawFile, previewFile, size)
	}

	args := []string{
		"-thumbnail", fmt.Sprintf("x%d", size),
		"-background", "white",
		//"-alpha", "remove",
		"-colorspace", "RGB",
		rawFile + fmt.Sprintf("[%d]", page),
		previewFile,
	}

	log.Context(ctx).Infof("call imagick: '%s'", args)
	return callImagick(args...)
}

func generatePicture(ctx context.Context, rawFile string, pictureFile string) error {
	args := []string{
		"-density", "300",
		rawFile,
		"-depth", "8",
		pictureFile,
	}
	log.Context(ctx).Infof("call imageick: '%s'", args)
	return callImagick(args...)
}
