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

package config

import (
	"fmt"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"io"
	"os"
	"path"
	"sync"
)

func InitLogging() error {

	filemode := os.O_APPEND | os.O_CREATE | os.O_WRONLY
	fileperm := os.FileMode(0760)

	if C.Logging.LogDirectory == "" {
		C.Logging.LogDirectory = "log"
	}
	dir := C.Logging.LogDirectory
	err := os.MkdirAll(C.Logging.LogDirectory, os.ModePerm)
	if err != nil {
		return fmt.Errorf("create log dir: %v", err)
	}

	httpFilename := path.Join(dir, C.Logging.HttpLogFile)
	logFilename := path.Join(dir, C.Logging.LogFile)

	formatter := &prefixed.TextFormatter{
		ForceColors:      false,
		DisableColors:    true,
		ForceFormatting:  false,
		DisableTimestamp: false,
		DisableUppercase: false,
		FullTimestamp:    false,
		TimestampFormat:  "",
		DisableSorting:   false,
		QuoteEmptyFields: false,
		QuoteCharacter:   "",
		SpacePadding:     0,
		Once:             sync.Once{},
	}

	logrus.SetFormatter(formatter)

	logLevel, err := logrus.ParseLevel(C.Logging.Loglevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %s", C.Logging.Loglevel)
	}
	logrus.SetLevel(logLevel)

	if C.Logging.LogHttp {
		httpLogFile, err := os.OpenFile(httpFilename, filemode, fileperm)
		if err != nil {
			return fmt.Errorf("open http log file: %v", err)
		}
		C.Logging.httpLog = httpLogFile
		C.Logging.HttpLog = logrus.New()

		if C.Logging.LogStdout {
			writer := io.MultiWriter(httpLogFile, os.Stdout)
			C.Logging.HttpLog.SetOutput(writer)
		} else {
			C.Logging.HttpLog.SetOutput(httpLogFile)
		}
		C.Logging.HttpLog.SetLevel(logrus.InfoLevel)
	}

	logFile, err := os.OpenFile(logFilename, filemode, fileperm)
	if err != nil {
		return fmt.Errorf("open log file: %v", err)
	}
	C.Logging.log = logFile

	if C.Logging.LogStdout {
		writer := io.MultiWriter(logFile, os.Stdout)
		logrus.SetOutput(writer)
	} else {
		logrus.SetOutput(logFile)
	}
	return nil
}

func DeinitLogging() {
	var logErr error
	var httpErr error
	if C.Logging.log != nil {
		logrus.SetOutput(os.Stdout)
		logErr = C.Logging.log.Close()
		C.Logging.log = nil
	}
	if C.Logging.httpLog != nil {
		httpErr = C.Logging.httpLog.Close()
		C.Logging.HttpLog.SetOutput(os.Stdout)
		C.Logging.httpLog = nil
	}

	if logErr == nil && httpErr == nil {
		return
	}

	logrus.Errorf("log errors: %v, %v", logErr, httpErr)
}
