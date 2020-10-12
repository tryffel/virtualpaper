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
	"net/http"
)

func respJson(resp http.ResponseWriter, body interface{}, statusCode int) error {
	var err error
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

func respOk(resp http.ResponseWriter, body interface{}) error {
	return respJson(resp, body, http.StatusOK)
}

func respBadRequest(resp http.ResponseWriter, reason string, body interface{}) error {
	if reason != "" && body == nil {
		body = map[string]string{
			"Error": reason,
		}
	}
	return respJson(resp, body, http.StatusBadRequest)
}

func respInternalError(resp http.ResponseWriter) error {
	return respJson(resp, nil, http.StatusInternalServerError)
}
