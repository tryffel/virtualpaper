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
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

type MetadataRequest struct {
	KeyId   int `valid:"required" json:"key_id"`
	ValueId int `valid:"required" json:"value_id"`
}

type MetadataKeyRequest struct {
	Key     string `json:"key" valid:"required"`
	Comment string `json:"comment" valid:"-"`
}

type MetadataValueRequest struct {
	// Value of new metadata
	Value string `json:"value" valid:"required"`
	// Optional comment
	Comment string `json:"comment" valid:"-"`
	// MatchDocuments instructs to try to match documents for this value.
	MatchDocuments bool   `json:"match_documents" valid:"-"`
	MatchType      string `json:"match_type" valid:"metadata_rule_type"`
	MatchFilter    string `json:"match_filter" valid:"-"`
}

type metadataUpdateRequest struct {
	Metadata []MetadataRequest `valid:"required" json:"metadata"`
}

func (m *metadataUpdateRequest) toMetadataArray() []*models.Metadata {
	metadata := make([]*models.Metadata, len(m.Metadata))

	for i, v := range m.Metadata {
		metadata[i] = v.toMetadata()
	}
	return metadata
}

func (m MetadataRequest) toMetadata() *models.Metadata {
	return &models.Metadata{
		KeyId:   m.KeyId,
		ValueId: m.ValueId,
	}
}

func (a *Api) getMetadataKeys(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/metadata/keys Metadata GetMetadataKeys
	// Get metadata keys
	// Responses:
	//  200: MetadataKeyResponse
	handler := "Api.getMetadataKeys"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	paging, err := getPaging(req)
	if err != nil {
		logrus.Warningf("invalid paging: %v", err)
		paging.Limit = 100
		paging.Offset = 0
	}

	sort, err := getSortParams(req, &models.MetadataKey{})
	if err != nil {
		respError(resp, err, handler)
		return
	}

	var sortfield = "name"
	var sortOrder = true

	if len(sort) > 0 {
		sortfield = sort[0].Key
		sortOrder = sort[0].Order
	}

	filter, err := getMetadataFilter(req)
	if err != nil {
		respError(resp, fmt.Errorf("get metadata filter: %v", err), handler)
		return
	}

	keys, err := a.db.MetadataStore.GetKeys(user, filter, storage.NewSortKey(sortfield, "name", sortOrder),
		paging)
	if err != nil {
		respError(resp, err, handler)
	}

	respResourceList(resp, keys, len(*keys))
}

func (a *Api) getMetadataKey(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/metadata/keys/{id} Metadata GetMetadataKey
	// Get metadata key
	// Responses:
	//  200: MetadataKeyResponse
	handler := "Api.getMetadataKey"
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

	key, err := a.db.MetadataStore.GetKey(user, id)
	if err != nil {
		respError(resp, err, handler)
	}

	respResourceList(resp, key, 1)
}

func (a *Api) getMetadataKeyValues(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/metadata/keys/{id}/values Metadata GetMetadataKeyValues
	// Get metadata key values
	// Responses:
	//  200: MetadataKeyValueResponse
	handler := "Api.getMetadataKeyValues"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	key, err := getParamIntId(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	paging, err := getPaging(req)
	if err != nil {
		logrus.Warningf("invalid paging: %v", err)
		paging.Limit = 100
		paging.Offset = 0
	}

	sort, err := getSortParams(req, &models.MetadataValue{})
	if err != nil {
		respError(resp, err, handler)
		return
	}

	var sortfield = "name"
	var sortOrder = true

	if len(sort) > 0 {
		sortfield = sort[0].Key
		sortOrder = sort[0].Order
	}

	keys, err := a.db.MetadataStore.GetValues(user, key, storage.NewSortKey(sortfield, "name", sortOrder), paging)
	if err != nil {
		respError(resp, err, handler)
	}

	respResourceList(resp, keys, len(*keys))
}

func (a *Api) updateDocumentMetadata(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/documents/{id}/metadata Documents UpdateDocumentMetadata
	// Update document metadata
	// Responses:
	//  200: DocumentResponse
	handler := "Api.updateDocumentMetadata"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	documentId := getParamId(req)

	dto := &metadataUpdateRequest{}
	err := unMarshalBody(req, dto)
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

func (a *Api) addMetadataKey(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/metadata/keys Metadata AddMetadataKey
	// Add metadata key
	// Responses:
	//  200: MetadataKeyResponse
	handler := "Api.addMetadataKey"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	dto := &MetadataKeyRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	key := &models.MetadataKey{
		UserId:    user,
		Key:       dto.Key,
		CreatedAt: time.Now(),
		Comment:   dto.Comment,
	}

	err = a.db.MetadataStore.CreateKey(user, key)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	respResourceList(resp, key, 1)
}

func (a *Api) addMetadataValue(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/metadata/keys/{id}/values Metadata AddMetadataKeyValues
	// Add metadata key values
	// Responses:
	//  200: MetadataKeyValueResponse
	//  400: String
	handler := "Api.addMetadataValue"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	keyId, err := getParamIntId(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	dto := &MetadataValueRequest{}
	err = unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	value := &models.MetadataValue{
		UserId:         user,
		KeyId:          keyId,
		Value:          dto.Value,
		CreatedAt:      time.Now(),
		Comment:        dto.Comment,
		MatchDocuments: dto.MatchDocuments,
		MatchType:      models.MetadataRuleType(dto.MatchType),
		MatchFilter:    dto.MatchFilter,
	}

	if value.MatchType == "" {
		value.MatchType = models.MetadataMatchExact
	}
	if value.MatchType != models.MetadataMatchRegex && value.MatchType != models.MetadataMatchExact {
		err := errors.ErrInvalid
		err.ErrMsg = "match type must be either exact or regex"
		respError(resp, err, handler)
		return
	}

	err = a.db.MetadataStore.CreateValue(user, value)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	respResourceList(resp, value, 1)
}

func (a *Api) updateMetadataValue(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/metadata/keys/{id}/values Metadata AddMetadataKeyValues
	// Add metadata key values
	// Responses:
	//  200: MetadataKeyValueResponse
	handler := "Api.updateMetadataValue"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	keyId, err := getParamInt(req, "key_id")
	if err != nil {
		respError(resp, err, handler)
		return
	}

	valueId, err := getParamInt(req, "value_id")
	if err != nil {
		respError(resp, err, handler)
		return
	}

	dto := &MetadataValueRequest{}
	err = unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	value := &models.MetadataValue{
		Id:             valueId,
		UserId:         user,
		KeyId:          keyId,
		Value:          dto.Value,
		Comment:        dto.Comment,
		MatchDocuments: dto.MatchDocuments,
		MatchType:      models.MetadataRuleType(dto.MatchType),
		MatchFilter:    dto.MatchFilter,
	}

	ownerShip, err := a.db.MetadataStore.UserHasKeyValue(user, keyId, valueId)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if !ownerShip {
		err := errors.ErrRecordNotFound
		respError(resp, err, handler)
		return
	}

	err = a.db.MetadataStore.UpdateValue(value)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	respResourceList(resp, value, 1)
}
