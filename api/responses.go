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
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/errors"
)

// PrettyTime prints time as default time string when marshaled as json
type PrettyTime time.Time

func (p PrettyTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(p).String() + "\""), nil
}

// formatValidatorError returns user-friendly presentation of govalidator.Error.
// if error is not govalidator.Error, return original error
func formatValidatorError(err error) string {
	if validatorErr, ok := err.(govalidator.Errors); ok {
		invalidFields := "invalid attribute: "
		multipleInvalid := false
		for _, v := range validatorErr.Errors() {
			if e, ok := v.(govalidator.Error); ok {
				if multipleInvalid {
					invalidFields += ", "
				}
				multipleInvalid = true
				invalidFields += e.Name
			}
		}
		return invalidFields
	}
	return err.Error()
}

func respJson(resp http.ResponseWriter, body interface{}, statusCode int) error {
	var err error
	if statusCode == 200 && resp.Header().Get("Cache-Control") == "" {
		resp.Header().Set("Cache-Control", "max-age=10")
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	if body != nil {
		err = json.NewEncoder(resp).Encode(body)
	}

	return err
}

func respUnauthorized(resp http.ResponseWriter) error {
	return respJson(resp, map[string]string{"Error": "Unauthorized"}, http.StatusUnauthorized)
}

func respOk(resp http.ResponseWriter, body interface{}) {
	err := respJson(resp, body, http.StatusOK)
	if err != nil {
		logrus.Errorf("write resp ok: %v", err)
	}
}

func respResourceList(resp http.ResponseWriter, body interface{}, totalCount int) {
	resp.Header().Set("Access-Control-Expose-Headers", "content-range")
	resp.Header().Set("Content-Range", strconv.Itoa(totalCount))
	err := respJson(resp, body, http.StatusOK)
	if err != nil {
		logrus.Errorf("send resource list resp: %v", err)
	}
}

func resourceList(c echo.Context, body interface{}, total int) error {
	c.Response().Header().Set("Access-Control-Expose-Headers", "content-range")
	c.Response().Header().Set("Content-Range", strconv.Itoa(total))
	return c.JSON(http.StatusOK, body)
}

func respBadRequest(resp http.ResponseWriter, reason string, body interface{}) {
	if reason != "" && body == nil {
		body = map[string]string{
			"Error": reason,
		}
	}
	err := respJson(resp, body, http.StatusBadRequest)
	if err != nil {
		logrus.Errorf("send resp bad request: %v", err)
	}
}

func respInternalError(resp http.ResponseWriter) {
	err := respJson(resp, nil, http.StatusInternalServerError)
	if err != nil {
		logrus.Errorf("write response: %v", err)
	}
}

func respForbiddenV2(error ...interface{}) error {
	return echo.NewHTTPError(http.StatusForbidden, error)
}

func respInternalErrorV2(error ...interface{}) error {
	return echo.NewHTTPError(http.StatusInternalServerError, error)
}

func httpErrorHandler(err error, c echo.Context) {
	var statuscode int
	var reason string
	if err == nil {
		// some middlware returned error
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		statuscode = he.Code
		reason = fmt.Sprintf("%v", he.Message)
	} else {
		appError, ok := err.(errors.Error)
		if ok {
			if appError.ErrType == errors.ErrAlreadyExists.ErrType {
				statuscode = http.StatusNotModified
				reason = appError.ErrMsg
			} else if appError.ErrType == errors.ErrRecordNotFound.ErrType {
				statuscode = http.StatusNotFound
				reason = appError.ErrMsg
			} else if appError.ErrType == errors.ErrInternalError.ErrType {
				statuscode = http.StatusInternalServerError
				reason = "internal error"
				//logrus.WithField("handler", handler).Errorf("internal error: %v", appError.ErrMsg)
				c.Logger().Error("internal error", appError.ErrMsg)
			} else if appError.ErrType == errors.ErrForbidden.ErrType {
				statuscode = http.StatusForbidden
				reason = appError.ErrMsg
			} else if appError.ErrType == errors.ErrUnauthorized.ErrType {
				statuscode = http.StatusUnauthorized
				reason = appError.ErrMsg
			} else if appError.ErrType == errors.ErrInvalid.ErrType {
				statuscode = http.StatusBadRequest
				reason = appError.ErrMsg
			} else {
				statuscode = http.StatusInternalServerError
				reason = "internal error, please try again shortly"
				logrus.Errorf("internal error: %v", err)
			}
		} else {
			statuscode = http.StatusInternalServerError
			reason = "internal error"
			c.Logger().Errorf("internal error: %v", err.Error())
		}
	}

	body := map[string]string{
		"Error": reason,
	}
	c.JSON(statuscode, body)
}
