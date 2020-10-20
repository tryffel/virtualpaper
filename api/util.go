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
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

func getUserId(req *http.Request) (int, bool) {
	ctx := req.Context()
	userId := ctx.Value("user_id")
	id, ok := userId.(int)
	return id, ok
}

func getParamId(req *http.Request) (int, error) {
	idStr := mux.Vars(req)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return -1, fmt.Errorf("id not integer")
	}

	if id < 0 {
		return -1, fmt.Errorf("id must be >0")
	}

	return id, err
}

// getSortParams parses and validates sorting parameters. This always returns list of params
// in the order of model.SortAttributes() and not in order of request.
func getSortParams(req *http.Request, model models.Modeler) ([]storage.SortKey, error) {
	sortKeys := make([]storage.SortKey, 0)

	err := req.ParseForm()
	if err != nil {
		return sortKeys, err
	}

	if len(req.Form) == 0 {
		return sortKeys, nil
	}

	vars := req.Form

	for _, v := range model.SortAttributes() {
		order := vars.Get(v)
		switch strings.ToUpper(order) {
		case "ASC":
			sort := storage.SortKey{
				Key:   v,
				Order: false,
			}
			sortKeys = append(sortKeys, sort)
		case "DESC":
			sort := storage.SortKey{
				Key:   v,
				Order: true,
			}
			sortKeys = append(sortKeys, sort)
		}
	}
	return sortKeys, nil
}
