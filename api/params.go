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
	"net/http"
	"strconv"
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
	return paging, nil
}
