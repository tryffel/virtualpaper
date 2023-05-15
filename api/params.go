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
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

type pageParams struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func bindPaging(c echo.Context) (storage.Paging, error) {
	params := &pageParams{
		Page:     1,
		PageSize: 50,
	}
	err := (&echo.DefaultBinder{}).BindQueryParams(c, params)
	if err != nil {
		return storage.Paging{Limit: 1}, echo.NewHTTPError(http.StatusBadRequest, "invalid paging: page and page_size must be numeric and > 0")
	}

	paging := storage.Paging{
		Offset: (params.Page - 1) * params.PageSize,
		Limit:  params.PageSize,
	}
	paging.Validate()
	return paging, nil
}

func bindPathId(c echo.Context) string {
	return c.Param("id")
}

func bindPathIdInt(c echo.Context) (int, error) {
	return bindPathInt(c, "id")
}

func bindPathInt(c echo.Context, name string) (int, error) {
	idStr := c.Param(name)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = name + " not integer"
		return -1, e
	}

	if id < 0 {
		e := errors.ErrInvalid
		e.ErrMsg = name + " must be >0"
		return -1, e
	}
	return id, err

}

type DocumentFilter struct {
	Query    string `json:"q" valid:"-"`
	Tag      string `json:"tag" valid:"-"`
	After    int64  `json:"after" valid:"-"`
	Before   int64  `json:"before" valid:"-"`
	Metadata string `json:"metadata" valid:"-"`
}

func getDocumentFilter(req *http.Request) (*search.DocumentFilter, error) {

	body := &DocumentFilter{}

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

func getMetadataFilter(req *http.Request) ([]int, error) {
	type MetadataFilter struct {
		Id []int
	}

	query := req.FormValue("filter")
	if query == "" || query == "{}" {
		return nil, nil
	}
	body := &MetadataFilter{}
	err := json.Unmarshal([]byte(query), body)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = "invalid json in search params"
		return nil, e
	}

	return body.Id, nil
}
