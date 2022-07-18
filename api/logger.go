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

package api

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
)

type LoggingWriter struct {
	resp   http.ResponseWriter
	status int
	length int
}

func (l *LoggingWriter) WriteHeader(status int) {
	l.length = 0
	l.status = status
	l.resp.WriteHeader(status)
}

func (l *LoggingWriter) Write(b []byte) (int, error) {
	l.length = len(b)
	if l.status == 0 {
		l.status = 200
	}
	return l.resp.Write(b)
}

func (l *LoggingWriter) Header() http.Header {
	return l.resp.Header()
}

// LogginMiddlware Provide logging for requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		logger := &LoggingWriter{
			resp: w,
		}
		next.ServeHTTP(logger, r)

		duration := time.Since(start).String()
		verb := r.Method
		url := r.RequestURI

		fields := make(map[string]interface{})
		fields["verb"] = verb
		fields["request"] = url
		fields["duration"] = duration
		fields["status"] = logger.status
		fields["length"] = logger.length

		if config.C.Logging.HttpLog != nil {
			config.C.Logging.HttpLog.WithFields(fields).Infof("http")
		} else {
			logrus.WithFields(fields).Infof("http")
		}
	})
}

func logCrudOp(resource string, action string, userId int, success *bool) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"module":   "api",
		"userid":   userId,
		"resource": resource,
		"action":   action,
		"success":  *success,
	})
}

func logCrudMetadata(userId int, action string, success *bool, fmt string, args ...interface{}) {
	logCrudOp("metadata", action, userId, success).Infof(fmt, args...)
}
