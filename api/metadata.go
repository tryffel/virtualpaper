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

type metadataRequest struct {
	KeyId   int `valid:"required" json:"key_id"`
	ValueId int `valid:"required" json:"value_id"`
}

type metadataUpdateRequest struct {
	Metadata []metadataRequest `valid:"required" json:"metadata"`
}

func (m *metadataUpdateRequest) toMetadataArray() []*models.Metadata {
	metadata := make([]*models.Metadata, len(m.Metadata))

	for i, v := range m.Metadata {
		metadata[i] = v.toMetadata()
	}
	return metadata
}

func (m metadataRequest) toMetadata() *models.Metadata {
	return &models.Metadata{
		KeyId:   m.KeyId,
		ValueId: m.ValueId,
	}
}

func (a *Api) getMetadataKeys(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getMetadataKeys"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	keys, err := a.db.MetadataStore.GetKeys(user)
	if err != nil {
		respError(resp, err, handler)
	}

	respResourceList(resp, keys, len(*keys))
}

func (a *Api) getMetadataKeyValues(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getMetadataKeyValues"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	key, err := getParamId(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	keys, err := a.db.MetadataStore.GetValues(user, key)
	if err != nil {
		respError(resp, err, handler)
	}

	respResourceList(resp, keys, len(*keys))
}

func (a *Api) updateDocumentMetadata(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.updateDocumentMetadata"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	documentId, err := getParamId(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	dto := &metadataUpdateRequest{}
	err = unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	metadata := dto.toMetadataArray()
	err = a.db.MetadataStore.UpdateDocumentKeyValues(user, documentId, metadata)
	if err != nil {
		respError(resp, err, handler)
	}
	respOk(resp, nil)
}
