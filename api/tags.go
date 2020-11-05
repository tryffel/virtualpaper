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
	"github.com/sirupsen/logrus"
	"net/http"
	"tryffel.net/go/virtualpaper/models"
)

type TagRequest struct {
	Key     string `valid:"required" json:"key"`
	Comment string `valid:"-" json:"comment"`
}

func (a *Api) getTags(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getTags"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	paging, err := getPaging(req)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
	}

	tags, n, err := a.db.MetadataStore.GetTags(user, paging)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	respResourceList(resp, tags, n)
}

func (a *Api) getTag(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getTag"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	id, err := getParamIntId(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	tag, err := a.db.MetadataStore.GetTag(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	respResourceList(resp, tag, 1)
}

func (a *Api) createTag(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.createTag"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	dto := &TagRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	tag := &models.Tag{
		Key:     dto.Key,
		Comment: dto.Comment,
	}

	err = a.db.MetadataStore.CreateTag(user, tag)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	respResourceList(resp, tag, 1)
}
