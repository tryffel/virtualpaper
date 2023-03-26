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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
)

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

func logCrudDocument(userId int, action string, success *bool, fmt string, args ...interface{}) {
	logCrudOp("document", action, userId, success).Infof(fmt, args...)
}

func logCrudRule(userId int, action string, success *bool, fmt string, args ...interface{}) {
	logCrudOp("processing-rule", action, userId, success).Infof(fmt, args...)
}

func logCrudAdminUsers(userId int, action string, success *bool, fmt string, args ...interface{}) {
	logCrudOp("admin-users", action, userId, success).Infof(fmt, args...)
}

func loggingMiddlware() echo.MiddlewareFunc {
	var logger *logrus.Logger

	if config.C.Logging.HttpLog != nil {
		logger = config.C.Logging.HttpLog
	} else {
		logger = logrus.StandardLogger()
	}

	logFunc := func(c echo.Context, values middleware.RequestLoggerValues) error {
		logger.WithFields(logrus.Fields{
			"method":    values.Method,
			"uri":       values.URIPath,
			"status":    values.Status,
			"requestId": values.RequestID,
		}).Info("request")
		return nil
	}

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper:          middleware.DefaultSkipper,
		BeforeNextFunc:   nil,
		LogValuesFunc:    logFunc,
		LogLatency:       false,
		LogProtocol:      false,
		LogRemoteIP:      true,
		LogHost:          false,
		LogMethod:        true,
		LogURI:           false,
		LogURIPath:       true,
		LogRoutePath:     false,
		LogRequestID:     true,
		LogReferer:       false,
		LogUserAgent:     false,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: false,
		LogResponseSize:  true,
		LogHeaders:       nil,
		LogQueryParams:   nil,
		LogFormValues:    nil,
	})
}
