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
	"github.com/sirupsen/logrus"
	"os/exec"
	"regexp"
	"tryffel.net/go/virtualpaper/config"
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
	logrus.Infof("found imagemagick version %s", ver)
	return ver
}
