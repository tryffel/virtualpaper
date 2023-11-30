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
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"tryffel.net/go/virtualpaper/models/aggregates"

	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type MetadataRequest struct {
	KeyId   int `valid:"required" json:"key_id"`
	ValueId int `valid:"required" json:"value_id"`
}

type MetadataKeyRequest struct {
	Key     string `json:"key" valid:"required,metadata,stringlength(1|30)"`
	Comment string `json:"comment" valid:"maxstringlength(1000),optional"`
	Icon    string `json:"icon" valid:"optional"`
	Style   string `json:"style" valid:"json,optional"`
}

type MetadataValueRequest struct {
	// Value of new metadata
	Value string `json:"value" valid:"required,metadata,stringlength(1|30)"`
	// Optional comment
	Comment string `json:"comment" valid:"maxstringlength(1000),optional"`
	// MatchDocuments instructs to try to match documents for this value.
	MatchDocuments bool `json:"match_documents" valid:"-"`
	// validate MatchType when creating, allowing default to be empty string
	MatchType   string `json:"match_type" valid:"in(regex|exact),optional"`
	MatchFilter string `json:"match_filter" valid:"maxstringlength(100),optional"`
}

type MetadataUpdateRequest struct {
	Metadata []MetadataRequest `valid:"required" json:"metadata"`
}

func (m *MetadataUpdateRequest) ToAggregate() []aggregates.Metadata {
	metadata := make([]aggregates.Metadata, len(m.Metadata))

	for i, v := range m.Metadata {
		metadata[i] = aggregates.Metadata{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}
	return metadata
}

func (m *MetadataUpdateRequest) UniqueKeys() []int {
	keyMap := map[int]bool{}
	for _, v := range m.Metadata {
		keyMap[v.KeyId] = true
	}

	keys := make([]int, len(keyMap))
	index := 0
	for i, _ := range keyMap {
		keys[index] = i
		index += 1
	}
	return keys
}

func (m *MetadataUpdateRequest) Keys() []int {
	keys := make([]int, len(m.Metadata))
	for i, v := range m.Metadata {
		keys[i] = v.KeyId
	}
	return keys
}

func (m *MetadataUpdateRequest) ToMetadataArray() []models.Metadata {
	metadata := make([]models.Metadata, len(m.Metadata))

	for i, v := range m.Metadata {
		metadata[i] = v.ToMetadata()
	}
	return metadata
}

func (m MetadataRequest) ToMetadata() models.Metadata {
	return models.Metadata{
		KeyId:   m.KeyId,
		ValueId: m.ValueId,
	}
}

func (a *Api) getMetadataKeys(c echo.Context) error {
	// swagger:route GET /api/v1/metadata/keys Metadata GetMetadataKeys
	// Get metadata keys
	// Responses:
	//  200: MetadataKeyResponse
	ctx := c.(UserContext)

	paging := getPagination(c)
	sort := getSort(c)
	if sort.Key == "" {
		sort.Key = "key"
	}

	filter, err := getMetadataFilter(c.Request())
	if err != nil {
		return err
	}

	keys, count, err := a.metadataService.GetKeys(getContext(ctx), ctx.UserId, filter, sort.ToKey(), paging.toPagination())
	if err != nil {
		return err
	}

	return resourceList(c, keys, count)
}

func (a *Api) getMetadataKey(c echo.Context) error {
	// swagger:route GET /api/v1/metadata/keys/{id} Metadata GetMetadataKey
	// Get metadata key
	// Responses:
	//  200: MetadataKeyResponse
	id, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	key, err := a.metadataService.GetKey(getContext(c), id)
	if err != nil {
		return err
	}
	return resourceList(c, key, 1)
}

func (a *Api) getMetadataKeyValues(c echo.Context) error {
	// swagger:route GET /api/v1/metadata/keys/{id}/values Metadata GetMetadataKeyValues
	// Get metadata key values
	// Responses:
	//  200: MetadataKeyValueResponse
	key, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	paging := getPagination(c)
	sort := getSort(c)
	keys, err := a.metadataService.GetKeyValues(getContext(c), key, sort.ToKey(), paging.toPagination())
	if err != nil {
		return err
	}

	return resourceList(c, keys, len(*keys))
}

func (a *Api) addMetadataKey(c echo.Context) error {
	// swagger:route POST /api/v1/metadata/keys Metadata AddMetadataKey
	// Add metadata key
	// Responses:
	//  200: MetadataKeyResponse
	ctx := c.(UserContext)

	opOk := false
	defer logCrudMetadata(ctx.UserId, "add key", &opOk, "")

	dto := &MetadataKeyRequest{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	key := &models.MetadataKey{
		UserId:    ctx.UserId,
		Key:       dto.Key,
		CreatedAt: time.Now(),
		Comment:   dto.Comment,
		Icon:      dto.Icon,
		Style:     dto.Style,
	}

	if key.Style == "" {
		key.Style = "{}"
	}

	err = a.metadataService.Create(getContext(c), ctx.UserId, key)
	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, key, 1)
}

func (a *Api) addMetadataValue(c echo.Context) error {
	// swagger:route POST /api/v1/metadata/keys/{id}/values Metadata AddMetadataKeyValues
	// Add metadata key values
	// Responses:
	//  200: MetadataKeyValueResponse
	//  400: String

	ctx := c.(UserContext)
	keyId, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	dto := &MetadataValueRequest{}
	err = unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudMetadata(ctx.UserId, "add value", &opOk, "key: %d", keyId)

	value := &models.MetadataValue{
		UserId:         ctx.UserId,
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
		return err
	}

	err = a.metadataService.CreateValue(getContext(c), value)
	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, value, 1)
}

func (a *Api) updateMetadataValue(c echo.Context) error {
	// swagger:route PUT /api/v1/metadata/keys/{id}/values Metadata UpdateMetadataKeyValue
	// UpdateJob metadata key value
	// Responses:
	//  200: MetadataKeyValueResponse

	ctx := c.(UserContext)
	keyId, err := bindPathInt(c, "keyId")
	if err != nil {
		return err
	}

	valueId, err := bindPathInt(c, "valueId")
	if err != nil {
		return err
	}

	dto := &MetadataValueRequest{}
	err = unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudMetadata(ctx.UserId, "update value", &opOk, "key: %d, value: %d", keyId, valueId)

	value := &models.MetadataValue{
		Id:             valueId,
		UserId:         ctx.UserId,
		KeyId:          keyId,
		Value:          dto.Value,
		Comment:        dto.Comment,
		MatchDocuments: dto.MatchDocuments,
		MatchType:      models.MetadataRuleType(dto.MatchType),
		MatchFilter:    dto.MatchFilter,
	}
	// rest should be enclodes in a transaction
	err = a.metadataService.UpdateValue(getContext(c), value)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, value, 1)
}

func (a *Api) updateMetadataKey(c echo.Context) error {
	// swagger:route PUT /api/v1/metadata/keys/{id} Metadata UpdateMetadataKeyValues
	// UpdateJob metadata key
	// Responses:
	//  200: MetadataKeyResponse

	ctx := c.(UserContext)

	keyId, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	dto := &MetadataKeyRequest{}
	err = unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudMetadata(ctx.UserId, "update key", &opOk, "key: %d", keyId)
	key := &models.MetadataKey{
		Id:      keyId,
		UserId:  ctx.UserId,
		Key:     dto.Key,
		Comment: dto.Comment,
		Icon:    dto.Icon,
		Style:   dto.Style,
	}

	if key.Style == "" {
		key.Style = "{}"
	}

	// rest should be enclosed in a transaction
	err = a.metadataService.UpdateKey(getContext(c), key)
	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, key, 1)
}

func (a *Api) deleteMetadataKey(c echo.Context) error {
	// swagger:route DELETE /api/v1/metadata/keys/{id} Metadata DeleteMetadataKey
	// Delete metadata key and all its values
	// Responses:
	//  200:

	ctx := c.(UserContext)
	keyId, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudMetadata(ctx.UserId, "delete key", &opOk, "key: %d", keyId)
	err = a.metadataService.DeleteKey(getContext(c), ctx.UserId, keyId)
	if err != nil {
		return err
	}
	opOk = true
	return c.String(http.StatusOK, "ok")
}

func (a *Api) deleteMetadataValue(c echo.Context) error {
	// swagger:route DELETE /api/v1/metadata/keys/{key_id}/value{id} Metadata DeleteMetadataValue
	// Delete metadata value
	// Responses:
	//  200:

	ctx := c.(UserContext)
	keyId, err := bindPathInt(c, "keyId")
	if err != nil {
		return err
	}

	valueId, err := bindPathInt(c, "valueId")
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudMetadata(ctx.UserId, "delete value", &opOk, "key: %d, value: %d", keyId, valueId)

	err = a.metadataService.DeleteValue(getContext(c), ctx.UserId, keyId, valueId)
	if err != nil {
		return err
	}
	opOk = true
	return c.String(http.StatusOK, "ok")
}

type linkedDocumentParams struct {
	// optional when document list is empty = clear linked documents
	DocumentIds []string `json:"documents" valid:"optional,uuidarray~Invalid ids"`
}

func (a *Api) getLinkedDocuments(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/linked-documents Metadata GetLinkedDocuments
	// Get linked documents
	// Responses:
	//  200:
	ctx := c.(UserContext)
	docId := c.Param("id")
	opOk := false
	defer logCrudMetadata(ctx.UserId, "get linked documents", &opOk, "document: %s", docId)
	docs, err := a.db.MetadataStore.GetLinkedDocuments(ctx.UserId, docId)
	if err != nil {
		return err
	} else {
		opOk = true
		return resourceList(c, docs, len(docs))
	}
}

func (a *Api) updateLinkedDocuments(c echo.Context) error {
	// swagger:route PUT /api/v1/documents/{id}/linked-documents Metadata UpdateLinkedDocuments
	// UpdateJob linked documents
	// Responses:
	//  200:

	ctx := c.(UserContext)
	docId := bindPathId(c)
	opOk := false
	defer logCrudMetadata(ctx.UserId, "update linked documents", &opOk, "document: %s", docId)

	dto := &linkedDocumentParams{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	if len(dto.DocumentIds) > 100 {
		e := errors.ErrInvalid
		e.ErrMsg = "Maximum number of linked documents is 100"
		return e
	}

	ownership, err := a.db.DocumentStore.UserOwnsDocuments(ctx.UserId, append(dto.DocumentIds, docId))
	if err != nil {
		return err
	}
	if !ownership {
		e := errors.ErrRecordNotFound
		e.ErrMsg = "document(s) not found"
		return e
	}

	err = a.db.MetadataStore.UpdateLinkedDocuments(ctx.UserId, docId, dto.DocumentIds)
	if err != nil {
		return err
	}

	docIds := make([]string, len(dto.DocumentIds)+1)
	docIds[0] = docId
	for i, _ := range dto.DocumentIds {
		docIds[i+1] = dto.DocumentIds[i]
	}

	err = a.db.DocumentStore.SetModifiedAt(docIds, time.Now())
	if err != nil {
		logrus.Errorf("update document updated_at when linking documents, docId: %s: %v", docId, err)
	}
	opOk = true

	data := map[string]string{"id": "1"}

	return c.JSON(http.StatusOK, data)
}
