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
	"net/http"
	"tryffel.net/go/virtualpaper/storage"
)

func unMarshalBody(r *http.Request, body interface{}) error {
	err := json.NewDecoder(r.Body).Decode(body)
	if err != nil {
		logrus.Debugf("invalid json: %v", err)
		return storage.ErrInvalid
	}

	ok, err := govalidator.ValidateStruct(body)
	if err != nil {
		return err
	}
	if !ok {
		return storage.ErrInvalid
	}
	return nil
}
