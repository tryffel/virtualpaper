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
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
)

type pageParams struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

func (p pageParams) toPagination() storage.Paging {
	return storage.Paging{
		Offset: (p.Page - 1) * p.PageSize,
		Limit:  p.PageSize,
	}
}

func mPagination() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			params := &pageParams{
				Page:     1,
				PageSize: 50,
			}
			err := (&echo.DefaultBinder{}).BindQueryParams(c, params)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid paging: page and page_size must be numeric and > 0")
			}
			params.PageSize = config.MaxRecords(params.PageSize)
			if ctx, ok := c.(UserContext); ok {
				ctx.pagination = *params
				return next(ctx)
			}
			ctx := Context{
				Context:    c,
				pagination: *params,
			}
			return next(ctx)
		}
	}
}

type SortKey struct {
	Key             string
	Order           bool
	CaseInsensitive bool
}

func (s SortKey) ToKey() storage.SortKey {
	return storage.SortKey{
		Key:             s.Key,
		Order:           s.Order,
		CaseInsensitive: s.CaseInsensitive,
	}
}

type SortKeys []SortKey

func getPagination(c echo.Context) pageParams {
	ctx, ok := c.(Context)
	if ok {
		return pageParams{
			Page:     1,
			PageSize: 20,
		}
	}
	userCtx, ok := c.(UserContext)
	if ok {
		return userCtx.pagination
	}
	return pageParams{
		Page:     ctx.pagination.Page,
		PageSize: ctx.pagination.PageSize,
	}
}

func mSort(model models.Modeler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			type sort struct {
				Sort  string `query:"sort"`
				Order string `query:"order"`
			}

			rawSort := &sort{}
			err := (&echo.DefaultBinder{}).BindQueryParams(c, rawSort)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid sort")
			}

			sortKeys := make([]storage.SortKey, 0)
			var sortVar string
			var sortOrder string
			if strings.HasPrefix(rawSort.Sort, "[") && strings.HasSuffix(rawSort.Sort, "]") {
				sortVar, sortOrder = parseSortParamArray(rawSort.Sort)
			} else {
				sortVar = rawSort.Sort
				sortOrder = rawSort.Order
			}

			for _, v := range model.SortAttributes() {
				if sortVar == v {
					caseInsensitive := false
					for _, sortKey := range model.SortNoCase() {
						if v == sortKey {
							caseInsensitive = true
							break
						}
					}
					switch strings.ToUpper(sortOrder) {
					case "ASC":
						sort := storage.NewSortKey(v, "id", false, caseInsensitive)
						sortKeys = append(sortKeys, sort)
					case "DESC":
						sort := storage.NewSortKey(v, "id", true, caseInsensitive)
						sortKeys = append(sortKeys, sort)
					default:
						sort := storage.NewSortKey(v, "id", false, caseInsensitive)
						sortKeys = append(sortKeys, sort)
					}
				}
			}
			if (len(sortKeys)) > 0 {
				if ctx, ok := c.(UserContext); ok {
					key := sortKeys[0]
					ctx.sort.Key = key.Key
					ctx.sort.Order = key.Order
					ctx.sort.CaseInsensitive = key.CaseInsensitive
					return next(ctx)
				}
			}
			return next(c)
		}
	}
}

func getSort(c echo.Context) SortKey {
	ctx, ok := c.(Context)
	if ok {
		return SortKey{
			Key:             "",
			Order:           true,
			CaseInsensitive: false,
		}
	}
	userCtx, ok := c.(UserContext)
	if ok {
		return userCtx.sort
	}
	return ctx.sort
}

func mDocumentOwner(service *services.DocumentService) func(idKey string) echo.MiddlewareFunc {
	return func(idKey string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.(UserContext)
				id := c.Param(idKey)
				owns, err := service.UserOwnsDocument(id, ctx.UserId)
				if err != nil {
					return err
				}
				if !owns {
					return echo.NewHTTPError(http.StatusNotFound, "not found")
				}
				return next(c)
			}
		}
	}
}

func mDocumentReadAccess(service *services.DocumentService) func(idKey string) echo.MiddlewareFunc {
	return func(idKey string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.(UserContext)
				id := c.Param(idKey)
				perms, err := service.DocumentPermissions(getContext(ctx), id, ctx.UserId)
				if err != nil {
					return err
				}
				if perms.Owner || perms.SharedPermissions.Read {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusNotFound, "not found")
			}
		}
	}
}

func mDocumentWriteAccess(service *services.DocumentService) func(idKey string) echo.MiddlewareFunc {
	return func(idKey string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.(UserContext)
				id := c.Param(idKey)
				perms, err := service.DocumentPermissions(getContext(ctx), id, ctx.UserId)
				if err != nil {
					return err
				}
				if perms.Owner || perms.SharedPermissions.Write {
					return next(c)
				}
				return echo.NewHTTPError(http.StatusNotFound, "not found")
			}
		}
	}
}

func mMetadataKeyOwner(service *services.MetadataService) func(idKey string) echo.MiddlewareFunc {
	return func(idKey string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.(UserContext)
				id, err := bindPathInt(c, idKey)
				if err != nil {
					userErr := errors.ErrInvalid
					userErr.ErrMsg = "id must be integer"
					return userErr
				}
				owns, err := service.UserOwnsKey(getContext(ctx), ctx.UserId, id)
				if err != nil {
					return err
				}
				if !owns {
					return echo.NewHTTPError(http.StatusNotFound, "not found")
				}
				return next(c)
			}
		}
	}
}

func mRuleOwner(service *services.RuleService) func(idKey string) echo.MiddlewareFunc {
	return func(idKey string) echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				ctx := c.(UserContext)
				id, err := bindPathInt(c, idKey)
				if err != nil {
					userErr := errors.ErrInvalid
					userErr.ErrMsg = "id must be integer"
					return userErr
				}
				owns, err := service.UserOwnsRule(getContext(ctx), ctx.UserId, id)
				if err != nil {
					return err
				}
				if !owns {
					return echo.NewHTTPError(http.StatusNotFound, "not found")
				}
				return next(c)
			}
		}
	}
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

func getMetadataIdFilter(req *http.Request) ([]int, error) {
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

func getMetadataSearchFilter(req *http.Request) (string, error) {
	type Query struct {
		Query string `json:"q"`
	}

	rawQuery := req.FormValue("filter")
	if rawQuery == "" || rawQuery == "{}" {
		return "", nil
	}

	query := Query{}
	err := json.Unmarshal([]byte(rawQuery), &query)
	if err != nil {
		e := errors.ErrInvalid
		e.ErrMsg = "invalid json in search params"
		return "", e
	}
	return query.Query, nil
}
