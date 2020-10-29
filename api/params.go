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
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

type pageParams struct {
	Page     int
	PageSize int
}

func getPaging(req *http.Request) (storage.Paging, error) {
	params := &pageParams{}

	pageStr := req.FormValue("page")
	if pageStr == "" {
		pageStr = "1"
	}
	sizeStr := req.FormValue("page_size")
	if sizeStr == "" {
		sizeStr = "10"
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}
	if page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = 10
	}

	if size < 1 {
		size = 5
	} else if size > 1000 {
		size = 1000
	}

	params.Page = page
	params.PageSize = size
	paging := storage.Paging{
		Offset: (params.Page - 1) * params.PageSize,
		Limit:  params.PageSize,
	}

	if paging.Offset < 0 {
		paging.Offset = 0
	}

	paging.Validate()
	return paging, nil
}

func getSearchQuery(req *http.Request) string {

	filter := req.FormValue("filter")
	if filter == "" {
		return ""
	}

	object := &map[string]interface{}{}
	err := json.Unmarshal([]byte(filter), object)
	if err != nil {
		logrus.Warningf("invalid filter: %v", err)
		return ""
	}

	val := (*object)["q"]
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func getDocumentFilter(req *http.Request) (*search.DocumentFilter, error) {
	type documentFilter struct {
		Query  string `json:"q"`
		Tag    string `json:"tag"`
		After  int64  `json:"after"`
		Before int64  `json:"before"`
	}

	body := &documentFilter{}

	query := req.FormValue("filter")
	if query == "" || query == "{}" {
		return nil, nil
	}
	err := json.Unmarshal([]byte(query), body)

	filter := &search.DocumentFilter{}
	filter.Query = body.Query
	filter.Tag = body.Tag
	if body.After != 0 {
		filter.After = time.Unix(body.After/1000, 0)
	}
	if body.Before != 0 {
		filter.Before = time.Unix(body.Before/1000, 0)
	}

	return filter, err
}
