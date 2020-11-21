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
	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"regexp"
	"tryffel.net/go/virtualpaper/errors"
)

func unMarshalBody(r *http.Request, body interface{}) error {
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		if err == io.EOF {
			return errors.ErrInvalid
		} else {
			logrus.Debugf("invalid json: %v", err)
			e := errors.ErrInvalid
			e.ErrMsg = "invalid json"
			return e
		}
		return errors.ErrInvalid
	}

	ok, err := govalidator.ValidateStruct(body)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = err.Error()
		return e
	}
	if !ok {
		e := errors.ErrInvalid
		e.ErrMsg = "invalid request"
		return e
	}
	return nil
}

var ipRegex = regexp.MustCompile("\\[?([\\d.:]+)\\]?:(\\d+)")

func getRemoteAddr(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	match := ipRegex.FindStringSubmatch(r.RemoteAddr)
	if len(match) == 3 {
		return match[1]
	}
	return r.RemoteAddr
}
