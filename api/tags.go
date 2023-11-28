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
	"github.com/labstack/echo/v4"
	"tryffel.net/go/virtualpaper/models"
)

type TagRequest struct {
	Key     string `valid:"required" json:"key"`
	Comment string `valid:"-" json:"comment"`
}

func (a *Api) getTags(c echo.Context) error {
	ctx := c.(UserContext)
	paging := getPagination(c)
	tags, n, err := a.db.MetadataStore.GetTags(ctx.UserId, paging.toPagination())
	if err != nil {
		return err
	}
	return resourceList(c, tags, n)
}

func (a *Api) getTag(c echo.Context) error {
	ctx := c.(UserContext)
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	tag, err := a.db.MetadataStore.GetTag(ctx.UserId, id)
	if err != nil {
		return err
	}
	return resourceList(c, tag, 1)
}

func (a *Api) createTag(c echo.Context) error {
	ctx := c.(UserContext)
	dto := &TagRequest{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	tag := &models.Tag{
		Key:     dto.Key,
		Comment: dto.Comment,
	}

	err = a.db.MetadataStore.CreateTag(ctx.UserId, tag)
	if err != nil {
		return err
	}

	return resourceList(c, tag, 1)
}
