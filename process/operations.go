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
	"github.com/sirupsen/logrus"
	"gopkg.in/gographics/imagick.v3/imagick"
	"strings"
)

func init() {

}

func generateThumbnail(rawFile string, previewFile string, page int, size int) error {
	imagick.Initialize()
	defer imagick.Terminate()

	logrus.Debugf("run 'convert -thumbnail'")

	args := []string{
		"convert",
		"-thumbnail", fmt.Sprintf("x%d", size),
		"-background", "white",
		"-alpha", "remove",
		rawFile + fmt.Sprintf("[%d]", page),
		previewFile,
	}

	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf("generate thumbnail: '%s'", strings.Join(args, " "))
	}

	msg, err := imagick.ConvertImageCommand(args)
	if err != nil {
		logrus.Errorf("imagick msg: %v", msg)
		return fmt.Errorf("imagick: %v", err)
	}
	return nil

}

func GetImagickVersion() string {
	ver, _ := imagick.GetVersion()
	return ver
}
