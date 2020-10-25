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
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"tryffel.net/go/virtualpaper/storage"
)

// PrettyTime prints time as default time string when marshaled as json
type PrettyTime time.Time

func (p PrettyTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + time.Time(p).String() + "\""), nil
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
		logrus.Error("write resp ok: %v", err)
	}
}

func respResourceList(resp http.ResponseWriter, body interface{}, totalCount int) {
	resp.Header().Set("Access-Control-Expose-Headers", "content-range")
	resp.Header().Set("Content-Range", strconv.Itoa(totalCount))
	err := respJson(resp, body, http.StatusOK)
	if err != nil {
		logrus.Error("send resource list resp: %v", err)
	}
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
		logrus.Error("write response: %v", err)
	}
}

func respError(resp http.ResponseWriter, err error, handler string) {
	var statuscode int
	var reason string

	if errors.Is(err, storage.ErrAlreadyExists) {
		statuscode = http.StatusNotModified
		reason = err.Error()
	} else if errors.Is(err, storage.ErrRecordNotFound) {
		statuscode = http.StatusNotFound
		reason = err.Error()
	} else if errors.Is(err, storage.ErrInternalError) {
		statuscode = http.StatusInternalServerError
		reason = "internal error"
		logrus.WithField("handler", handler).Errorf("internal error: %v", err)
	} else if errors.Is(err, storage.ErrForbidden) {
		statuscode = http.StatusForbidden
		reason = err.Error()
		logrus.Errorf("internal error: %v", err)
	} else if errors.Is(err, storage.ErrInvalid) {
		statuscode = http.StatusBadRequest
		reason = err.Error()
	} else {
		statuscode = http.StatusInternalServerError
		reason = "internal error, please try again shortly"
		logrus.Errorf("internal error: %v", err)
	}

	body := map[string]string{
		"Error": reason,
	}
	respErr := respJson(resp, body, statuscode)
	if respErr != nil {
		logrus.WithField("handler", handler).Errorf("write error response: %v", err)
	}
}
