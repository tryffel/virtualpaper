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
	"github.com/labstack/echo/v4"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/models/aggregates"
	"tryffel.net/go/virtualpaper/services"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/util/logger"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/storage"
)

func responseFromDocument(doc *models.Document) *aggregates.Document {
	resp := aggregates.DocumentToAggregate(doc)
	resp.PreviewUrl = fmt.Sprintf("%s/api/v1/documents/%s/preview", config.C.Api.PublicUrl, doc.Id)
	resp.DownloadUrl = fmt.Sprintf("%s/api/v1/documents/%s/download", config.C.Api.PublicUrl, doc.Id)
	return resp
}

func documentAggregateToResponse(doc *aggregates.Document) {
	doc.PreviewUrl = fmt.Sprintf("%s/api/v1/documents/%s/preview", config.C.Api.PublicUrl, doc.Id)
	doc.DownloadUrl = fmt.Sprintf("%s/api/v1/documents/%s/download", config.C.Api.PublicUrl, doc.Id)
}

type DocumentExistsResponse struct {
	Error string `json:"error"`
	Id    string `json:"id"`
	Name  string `json:"name"`
}

// DocumentUpdateRequest
// swagger:model DocumentUpdateRequestBody
type DocumentUpdateRequest struct {
	Name        string            `json:"name" valid:"required,stringlength(1|200)"`
	Description string            `json:"description" valid:"maxstringlength(1000),optional"`
	Filename    string            `json:"filename" valid:"optional"`
	Date        int64             `json:"date" valid:"optional,range(0|4106139691000)"` // year 2200 in ms
	Metadata    []MetadataRequest `json:"metadata" valid:"-"`
	Lang        string            `json:"lang" valid:"language, optional"`
}

func (a *Api) getDocuments(c echo.Context) error {
	// swagger:route GET /api/v1/documents Documents GetDocuments
	// Get documents
	//
	// responses:
	//   200: Document
	//handler := "Api.getDocuments"
	ctx := c.(UserContext)
	//user := ctx.User
	query, err := getDocumentFilter(c.Request())
	if err != nil {
		return err
	}

	if query != nil {
		return a.searchDocuments(ctx.UserId, query, c)
	}
	paging := getPagination(c)
	sort := getSort(c)
	docs, count, err := a.documentService.GetDocuments(ctx.UserId, paging.toPagination(), sort.ToKey(), true)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		return err
	}
	respDocs := make([]*aggregates.Document, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}
	return resourceList(c, respDocs, count)
}

func (a *Api) getDeletedDocuments(c echo.Context) error {
	// swagger:route GET /api/v1/documents/deleted Documents GetDeletedDocuments
	// Get deleted documents
	//
	// responses:
	//   200: Document
	ctx := c.(UserContext)

	paging := getPagination(c)
	sort := getSort(c)
	docs, count, err := a.documentService.GetDeletedDocuments(ctx.UserId, paging.toPagination(), sort.ToKey(), true)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		return err
	}
	respDocs := make([]*aggregates.Document, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}
	return resourceList(c, respDocs, count)
}

func (a *Api) getDocument(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id} Documents GetDocument
	// Get document
	// responses:
	//   200: Document

	ctx := c.(UserContext)
	id := c.Param("id")

	visit := c.QueryParam("visit")
	if visit != "1" && visit != "0" && visit != "" {
		err := errors.ErrInvalid
		err.ErrMsg = "query parameter 'visit' must be either 1 or 0"
		return err
	}

	doc, err := a.documentService.GetDocument(getContext(c), ctx.UserId, id, visit == "1")
	if err != nil {
		return err
	}
	documentAggregateToResponse(doc)
	return c.JSON(http.StatusOK, doc)
}

func (a *Api) getDocumentContent(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/content Documents GetDocumentContent
	// Get full document parsed content
	// responses:
	//   200: Document

	id := c.Param("id")
	content, err := a.documentService.GetContent(getContext(c), id)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, *content)
}

func (a *Api) getDocumentLogs(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/jobs Documents GetDocumentJobs
	// Get processing job history related to document
	// responses:
	//   200: Document
	id := c.Param("id")
	job, err := a.db.JobStore.GetJobsByDocumentId(id)
	if err != nil {
		return err
	}
	return resourceList(c, job, len(*job))
}

func (a *Api) getDocumentPreview(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/preview Documents GetDocumentPreview
	// Get document preview, a small png image of first page of document.
	// responses:

	id := c.Param("id")
	doc, err := a.db.DocumentStore.GetDocument(id)
	if err != nil {
		return err
	}

	filePath := storage.PreviewPath(doc.Id)
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		return err
	}
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := c.Response().Header()
	header.Set("Content-Type", "image/png")
	header.Set("Content-Length", strconv.Itoa(int(stat.Size())))
	header.Set("Content-Disposition", "attachment; filename="+doc.Id+".png")
	header.Set("Cache-Control", "max-age=600")

	defer file.Close()
	_, err = io.Copy(c.Response(), file)
	if err != nil {
		logrus.Errorf("send file over http: %v", err)
	}
	return nil
}

func (a *Api) uploadFile(c echo.Context) error {
	// swagger:route POST /api/v1/documents Documents UploadFile
	// Upload new document file. New document already contains id, name, filename and timestamps.
	// Otherwise document is not processed yet and lacks other fields.
	// Consumes:
	// - multipart/form-data
	//
	// Responses:
	//  200: Document
	//  400: DocumentExistsResponse
	ctx := c.(UserContext)
	var err error
	opOk := false
	documentId := ""

	defer func() {
		logCrudDocument(ctx.UserId, "upload", &opOk, "document: %s", documentId)
	}()

	req := c.Request()

	err = req.ParseMultipartForm(1024 * 1024 * 500)
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid form: %v", err)
		userError.Err = err
		return userError
	}
	formKey := req.FormValue("name")
	reader, header, err := req.FormFile(formKey)
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid file: %v", err)
		userError.Err = err
		return userError
	}

	sanitizedFormKey := govalidator.SafeFileName(formKey)
	name := govalidator.SafeFileName(header.Filename)
	mimetype := process.MimeTypeFromName(sanitizedFormKey)

	if mimetype == "application/octet-stream" {
		mimetype = "text/plain"
	}

	defer reader.Close()
	buf := make([]byte, 500)
	_, err = reader.Read(buf)
	if err != nil {
		return respInternalErrorV2(fmt.Errorf("peek file contents: %v", err))
	}

	_, err = reader.Seek(0, io.SeekStart)
	if err != nil {
		return respInternalErrorV2(fmt.Errorf("seek file to start: %v", err))
	}

	detectedFileType := http.DetectContentType(buf)
	if strings.HasPrefix(detectedFileType, "text/plain") {
		detectedFileType = "text/plain"
	}
	if detectedFileType != mimetype {
		logger.Context(c.Request().Context()).Warnf("uploaded document detected mimetype does not match reported, given %s, detected %s", header.Filename, detectedFileType)
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("illegal mimetype")
		userError.Err = err
		return userError
	}

	doc, err := a.documentService.UploadFile(c.Request().Context(), &services.UploadedFile{
		UserId:   ctx.UserId,
		Filename: name,
		Mimetype: mimetype,
		Size:     header.Size,
		File:     reader,
	})

	if errors.Is(err, errors.ErrAlreadyExists) && doc != nil {
		body := DocumentExistsResponse{
			Error: "document exists",
			Id:    doc.Id,
			Name:  doc.Name,
		}
		return c.JSON(http.StatusBadRequest, body)
	}

	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(http.StatusOK, responseFromDocument(doc))
}

func (a *Api) getEmptyDocument(resp http.ResponseWriter, req *http.Request) {
	doc := &models.Document{}
	respResourceList(resp, responseFromDocument(doc), 1)
}

func (a *Api) downloadDocument(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id} Documents DownloadDocument
	// Downloads original document
	// Responses:
	//  200: Document

	ctx := c.(UserContext)
	var err error
	id := c.Param("id")

	opOk := false
	defer logCrudDocument(ctx.UserId, "download", &opOk, "document: %s", id)
	file, err := a.documentService.DocumentFile(id)
	defer file.File.Close()

	resp := c.Response()
	resp.Header().Set("Content-Type", file.Mimetype)
	resp.Header().Set("Content-Length", strconv.Itoa(int(file.Size)))
	resp.Header().Set("Cache-Control", "max-age=600")

	_, err = io.Copy(resp, file.File)
	if err != nil {
		logrus.Errorf("send file over http: %v", err)
	}
	opOk = true
	return nil
}

func (a *Api) updateDocument(c echo.Context) error {
	// swagger:route PUT /api/v1/documents/{id} Documents UpdateDocument
	// Updates document
	// Responses:
	//  200: Document

	ctx := c.(UserContext)
	id := c.Param("id")
	dto := &DocumentUpdateRequest{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudDocument(ctx.UserId, "update", &opOk, "document: %s", id)
	dto.Filename = govalidator.SafeFileName(dto.Filename)
	metadata := make([]aggregates.Metadata, len(dto.Metadata))
	for i, v := range dto.Metadata {
		metadata[i] = aggregates.Metadata{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}

	doc := &aggregates.DocumentUpdate{
		Name:        dto.Name,
		Description: dto.Description,
		Filename:    dto.Filename,
		Date:        time.Time{},
		Metadata:    metadata,
		Lang:        dto.Lang,
	}

	if dto.Date != 0 {
		doc.Date = time.Unix(dto.Date/1000, 0)
	}

	updatedDoc, err := a.documentService.UpdateDocument(getContext(c), ctx.UserId, id, doc)
	opOk = err == nil
	if err != nil {
		return err
	}
	return resourceList(c, responseFromDocument(updatedDoc), 1)
}

func (a *Api) searchDocuments(userId int, filter *search.DocumentFilter, c echo.Context) error {
	paging := getPagination(c)
	sort := getSort(c)
	if filter.Sort == "id" {
		filter.Sort = ""
	}

	opOk := false
	defer logCrudDocument(userId, "search", &opOk, "metadata: %v, query: %v", filter.Metadata != "", filter.Query != "")

	res, n, err := a.documentService.SearchDocuments(userId, filter.Query, sort.ToKey(), paging.toPagination())
	if err != nil {
		return err
	}

	docs := make([]*aggregates.Document, len(res))
	for i, v := range res {
		docs[i] = responseFromDocument(v)
	}
	opOk = true
	return resourceList(c, docs, n)
}

func (a *Api) requestDocumentProcessing(c echo.Context) error {
	// swagger:route POST /api/v1/location Documents RequestProcessing
	// Request document re-processing
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)
	id := bindPathId(c)
	opOk := false
	defer logCrudDocument(ctx.UserId, "schedule processing", &opOk, "document: %s", id)
	err := a.documentService.RequestProcessing(getContext(c), ctx.UserId, id)
	opOk = err == nil
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, "")
}

func (a *Api) deleteDocument(c echo.Context) error {
	// swagger:route DELETE /api/v1/documents/:id Documents DeleteDocument
	// Delete document
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)
	id := c.Param("id")

	opOk := false
	defer logCrudDocument(ctx.UserId, "delete", &opOk, "document: %s", id)
	err := a.documentService.DeleteDocument(getContext(c), id, ctx.UserId)
	opOk = err == nil
	return err
}

type BulkEditDocumentsRequest struct {
	Documents      []string              `json:"documents" valid:"required"`
	AddMetadata    MetadataUpdateRequest `json:"add_metadata" valid:"-"`
	RemoveMetadata MetadataUpdateRequest `json:"remove_metadata" valid:"-"`
	Lang           string                `json:"lang" valid:"language, optional"`
	Date           int64                 `json:"date" valid:"optional,range(0|4106139691000)"` // year 2200 in ms
}

func (a *Api) bulkEditDocuments(c echo.Context) error {
	// swagger:route POST /api/v1/documents/bulkEdit Documents BulkEditDocuments
	// Edit multiple documents at once
	// consumes:
	//  - application/json
	//
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)
	dto := &BulkEditDocumentsRequest{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	if len(dto.RemoveMetadata.Metadata) == 0 && len(dto.AddMetadata.Metadata) == 0 && dto.Lang == "" && dto.Date == 0 {
		userErr := errors.ErrAlreadyExists
		userErr.ErrMsg = "no documents modified"
		return userErr
	}
	opOk := false
	defer logCrudDocument(ctx.UserId, "bulk edit", &opOk, "documents: %v, add metadata: %d, remove metadata: %d, set lang: '%s'",
		len(dto.Documents), len(dto.AddMetadata.Metadata), len(dto.RemoveMetadata.Metadata), dto.Lang)

	owns, err := a.db.DocumentStore.UserOwnsDocuments(ctx.UserId, dto.Documents)
	if err != nil {
		return err
	}
	if !owns {
		return respForbiddenV2()
	}

	req := aggregates.BulkEditDocumentsRequest{
		Documents:      dto.Documents,
		AddMetadata:    dto.AddMetadata.ToAggregate(),
		RemoveMetadata: dto.RemoveMetadata.ToAggregate(),
		Lang:           dto.Lang,
		Date:           dto.Date,
	}

	err = a.documentService.BulkEditDocuments(getContext(c), &req, ctx.UserId)
	opOk = err == nil
	if err != nil {
		return err
	}
	return resourceList(c, dto.Documents, len(dto.Documents))
}

type SearchSuggestRequest struct {
	Filter string `json:"filter" valid:"-"`
}

type SearchSuggestResponse struct {
	Suggestions []string `json:"suggestions"`
	Prefix      string   `json:"prefix"`
}

func (a *Api) searchSuggestions(c echo.Context) error {
	// swagger:route POST /api/v1/documents/search/suggest Documents SearchSuggestions
	// Get search suggestions
	// consumes:
	//  - application/json
	//
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError
	ctx := c.(UserContext)

	dto := &SearchSuggestRequest{}
	err := unMarshalBody(c.Request(), dto)
	if err != nil {
		return err
	}

	suggestions, err := a.search.SuggestSearch(ctx.UserId, dto.Filter)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, suggestions)
}

func (a *Api) getDocumentHistory(c echo.Context) error {
	// swagger:route GET /api/v1/documents/:id/history Documents GetHistory
	// Get document history
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	ctx := c.(UserContext)
	id := c.Param("id")
	opOk := false
	defer logCrudDocument(ctx.UserId, "delete", &opOk, "document: %s", id)
	data, err := a.documentService.GetHistory(getContext(c), ctx.UserId, id)
	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, data, len(*data))
}

func (a *Api) restoreDeletedDocument(c echo.Context) error {
	// swagger:route PUT /api/v1/documents/deleted/:id/restore User UserRestoreDeletedDocument
	// Restore deleted document
	//
	// responses:
	//   200: RespUserInfo

	ctx := c.(UserContext)
	docId := bindPathId(c)

	opOk := false
	defer func() {
		logCrudDocument(ctx.UserId, "restore deleted document", &opOk, "restore deleted document %s", docId)
	}()

	doc, err := a.documentService.RestoreDeletedDocument(getContext(c), docId, ctx.UserId)
	opOk = err == nil
	if err != nil {
		return err
	}
	return c.JSON(200, responseFromDocument(doc))
}

func (a *Api) flushDeletedDocument(c echo.Context) error {
	// swagger:route DELETE /api/v1/documents/deleted/:id User UserFlushDeletedDocument
	// Flush deleted document
	//
	// responses:
	//   200: RespUserInfo

	ctx := c.(UserContext)
	docId := bindPathId(c)

	opOk := false
	defer func() {
		logCrudDocument(ctx.UserId, "flush deleted document", &opOk, "flush deleted document %s", docId)
	}()

	err := a.documentService.FlushDeletedDocument(getContext(c), docId)
	opOk = err == nil
	return err
}
