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
	"errors"
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

func getUser(req *http.Request) (*models.User, bool) {
	ctx := req.Context()
	userId := ctx.Value("user")
	user, ok := userId.(*models.User)
	return user, ok
}

func userIsAdmin(req *http.Request) (bool, error) {
	user, ok := getUser(req)
	if !ok {
		return false, errors.New("user record not found in context")
	}
	return user.IsAdmin, nil
}

func getParamId(req *http.Request) (int, error) {
	idStr := mux.Vars(req)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		e := storage.ErrInvalid
		e.ErrMsg = "id not integer"
		return -1, e
	}

	if id < 0 {
		e := storage.ErrInvalid
		e.ErrMsg = "id must be >0"
		return -1, e
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

	sortVar := vars.Get("sort")
	sortOrder := vars.Get("order")

	if strings.HasPrefix(sortVar, "[") && strings.HasSuffix(sortVar, "]") {
		sortVar, sortOrder = parseSortParamArray(sortVar)
	}

	for _, v := range model.SortAttributes() {
		if sortVar == v {
			switch strings.ToUpper(sortOrder) {
			case "ASC":
				sort := storage.NewSortKey(v, "id", false)
				sortKeys = append(sortKeys, sort)
			case "DESC":
				sort := storage.NewSortKey(v, "id", false)
				sortKeys = append(sortKeys, sort)
			default:
				sort := storage.NewSortKey(v, "id", false)
				sortKeys = append(sortKeys, sort)
			}
		}
	}
	return sortKeys, nil
}

func parseSortParamArray(s string) (string, string) {
	s = strings.Trim(s, "[]\"")
	s = strings.Replace(s, "\"", "", 2)
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return s, ""
	}
	return parts[0], parts[1]
}
