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
	"net/http"
	"strconv"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

type pageParams struct {
	Page     int
	PageSize int
}

func getPaging(req *http.Request) (storage.Paging, error) {
	params := &pageParams{}

	var pageStr string
	var sizeStr string
	page := 1
	pageSize := 10

	pageStr = req.FormValue("page")
	if pageStr == "" {
		page = 1
	} else {
		gotPage, err := strconv.Atoi(pageStr)
		if err == nil && gotPage > 0 {
			page = gotPage
		}
	}

	sizeStr = req.FormValue("page_size")
	if sizeStr == "" {
		pageSize = 10
	} else {
		size, err := strconv.Atoi(sizeStr)
		if err == nil && size > 0 {
			if size > config.MaxRows {
				pageSize = config.MaxRows
			} else {
				pageSize = size
			}
		}
	}

	params.Page = page
	params.PageSize = pageSize
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

func getDocumentFilter(req *http.Request) (*search.DocumentFilter, error) {
	type documentFilter struct {
		Query    string `json:"q" valid:"-"`
		Tag      string `json:"tag" valid:"-"`
		After    int64  `json:"after" valid:"-"`
		Before   int64  `json:"before" valid:"-"`
		Metadata string `json:"metadata" valid:"-"`
	}

	body := &documentFilter{}

	query := req.FormValue("filter")
	if query == "" || query == "{}" {
		return nil, nil
	}
	err := json.Unmarshal([]byte(query), body)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = "invalid json in search params"
		return nil, e
	}

	ok, err := govalidator.ValidateStruct(body)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = formatValidatorError(err)
		return nil, e
	}
	if !ok {
		return nil, errors.ErrInvalid
	}

	filter := &search.DocumentFilter{}
	filter.Query = body.Query
	filter.Tag = body.Tag
	if body.After != 0 {
		filter.After = time.Unix(body.After/1000, 0)
	}
	if body.Before != 0 {
		filter.Before = time.Unix(body.Before/1000, 0)
	}
	filter.Metadata = body.Metadata

	return filter, err
}
