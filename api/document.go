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

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

// DocumentResponse
type DocumentResponse struct {
	// swagger:strfmt uuid
	Id          string            `json:"id"`
	Name        string            `json:"name"`
	Filename    string            `json:"filename"`
	Content     string            `json:"content"`
	Description string            `json:"description"`
	CreatedAt   int64             `json:"created_at"`
	UpdatedAt   int64             `json:"updated_at"`
	Date        int64             `json:"date"`
	PreviewUrl  string            `json:"preview_url"`
	DownloadUrl string            `json:"download_url"`
	Mimetype    string            `json:"mimetype"`
	Type        string            `json:"type"`
	Size        int64             `json:"size"`
	PrettySize  string            `json:"pretty_size"`
	Status      string            `json:"status"`
	Metadata    []models.Metadata `json:"metadata"`
	Tags        []models.Tag      `json:"tags"`
}

func responseFromDocument(doc *models.Document) *DocumentResponse {
	resp := &DocumentResponse{
		Id:          doc.Id,
		Name:        doc.Name,
		Filename:    doc.Filename,
		Content:     doc.Content,
		Description: doc.Description,
		CreatedAt:   doc.CreatedAt.Unix() * 1000,
		UpdatedAt:   doc.UpdatedAt.Unix() * 1000,
		Date:        doc.Date.Unix() * 1000,
		PreviewUrl:  fmt.Sprintf("%s/api/v1/documents/%s/preview", config.C.Api.PublicUrl, doc.Id),
		DownloadUrl: fmt.Sprintf("%s/api/v1/documents/%s/download", config.C.Api.PublicUrl, doc.Id),
		Mimetype:    doc.Mimetype,
		Type:        doc.GetType(),
		Size:        doc.Size,
		PrettySize:  doc.GetSize(),
		Metadata:    doc.Metadata,
		Tags:        doc.Tags,
	}
	return resp
}

type DocumentExistsResponse struct {
	Error string `json:"error"`
	Id    string `json:"id"`
	Name  string `json:"name"`
}

// DocumentUpdateRequest
// swagger:model DocumentUpdateRequestBody
type DocumentUpdateRequest struct {
	Name        string            `json:"name" valid:"-"`
	Description string            `json:"description" valid:"-"`
	Filename    string            `json:"filename" valid:"-"`
	Date        int64             `json:"date" valid:"-"`
	Metadata    []MetadataRequest `json:"metadata" valid:"-"`
}

func (a *Api) getDocuments(c echo.Context) error {
	// swagger:route GET /api/v1/documents Documents GetDocuments
	// Get documents
	//
	// responses:
	//   200: DocumentResponse
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

	paging, err := bindPaging(c)
	if err != nil {
		return err
	}

	sort, err := getSortParams(c.Request(), &models.Document{})
	if err != nil {
		return err
	}

	if len(sort) == 0 {
		sort = append(sort, storage.SortKey{})
	}

	docs, count, err := a.db.DocumentStore.GetDocuments(ctx.UserId, paging, sort[0], true)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		return err
	}
	respDocs := make([]*DocumentResponse, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}
	return resourceList(c, respDocs, count)
}

func (a *Api) getDocument(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id} Documents GetDocument
	// Get document
	// responses:
	//   200: DocumentResponse

	ctx := c.(UserContext)
	id := c.Param("id")
	doc, err := a.db.DocumentStore.GetDocument(ctx.UserId, id)
	if err != nil {
		return err
	}

	status, err := a.db.JobStore.GetDocumentStatus(doc.Id)
	if err != nil {
		return err
	}

	metadata, err := a.db.MetadataStore.GetDocumentMetadata(ctx.UserId, id)
	if err != nil {
		return err
	}
	doc.Metadata = *metadata

	tags, err := a.db.MetadataStore.GetDocumentTags(ctx.UserId, id)
	if err != nil {
		return err
	}
	doc.Tags = *tags

	respDoc := responseFromDocument(doc)
	respDoc.Status = status
	return c.JSON(http.StatusOK, respDoc)
}

func (a *Api) getDocumentContent(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/content Documents GetDocumentContent
	// Get full document parsed content
	// responses:
	//   200: DocumentResponse

	ctx := c.(UserContext)
	id := c.Param("id")

	content, err := a.db.DocumentStore.GetContent(ctx.UserId, id)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, *content)
}

func (a *Api) getDocumentLogs(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/jobs Documents GetDocumentJobs
	// Get processing job history related to document
	// responses:
	//   200: DocumentResponse

	ctx := c.(UserContext)
	id := c.Param("id")
	owns, err := a.db.DocumentStore.UserOwnsDocument(id, ctx.UserId)
	if err != nil {
		return err
	}

	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	job, err := a.db.JobStore.GetByDocument(id)
	if err != nil {
		return err
	}
	return resourceList(c, job, len(*job))
}

func (a *Api) getDocumentPreview(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id}/preview Documents GetDocumentPreview
	// Get document preview, a small png image of first page of document.
	// responses:

	ctx := c.(UserContext)
	id := c.Param("id")
	doc, err := a.db.DocumentStore.GetDocument(ctx.UserId, id)
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
	//  200: DocumentResponse
	//  400: DocumentExistsResponse
	ctx := c.(UserContext)
	var err error
	opOk := false
	documentId := ""

	defer func() {
		logCrudMetadata(ctx.UserId, "upload", &opOk, "document: %s", documentId)
	}()

	req := c.Request()

	err = req.ParseMultipartForm(1024 * 1024 * 500)
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid form: %v", err)
		userError.Err = err
		return userError
	}
	reader, header, err := req.FormFile("file")
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid file: %v", err)
		userError.Err = err
		return userError
	}

	mimetype := header.Header.Get("Content-Type")

	if mimetype == "application/octet-stream" {
		mimetype = "text/plain"
	}

	defer reader.Close()

	tempHash := config.RandomString(10)

	document := &models.Document{
		Id:       "",
		UserId:   ctx.UserId,
		Name:     header.Filename,
		Content:  "",
		Filename: header.Filename,
		Hash:     tempHash,
		Mimetype: mimetype,
		Size:     header.Size,
		Date:     time.Now(),
	}

	if !process.MimeTypeIsSupported(mimetype, header.Filename) {
		e := errors.ErrInvalid
		e.ErrMsg = fmt.Sprintf("unsupported file type: %v", header.Filename)
		req.Body.Close()
		return e
	}

	tempFileName := storage.TempFilePath(tempHash)
	inputFile, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		c.Logger().Errorf("open new file for saving upload: %v", err)
		//respError(resp, fmt.Errorf("open new file for saving upload: %v", err), handler)
		return err
	}
	n, err := inputFile.ReadFrom(reader)
	if err != nil {
		return fmt.Errorf("write uploaded file to disk: %v", err)
	}

	if n != header.Size {
		logrus.Warningf("did not fully read file: %d, got: %d", header.Size, n)
	}

	err = inputFile.Close()
	if err != nil {
		return fmt.Errorf("close file: %v", err)
	}

	hash, err := process.GetHash(tempFileName)
	if err != nil {
		return fmt.Errorf("get hash for temp file: %v", err)
	}

	existingDoc, err := a.db.DocumentStore.GetByHash(0, hash)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
		} else {
			return fmt.Errorf("get existing document by hash: %v", err)
		}
	}

	if existingDoc != nil {
		if existingDoc.Id != "" {
			body := DocumentExistsResponse{
				Error: "document exists",
				Id:    existingDoc.Id,
				Name:  existingDoc.Name,
			}
			err := os.Remove(tempFileName)
			if err != nil {
				c.Logger().Errorf("remove duplicated temp file: %v", err)
			}
			return c.JSON(http.StatusBadRequest, body)
		}
	}

	document.Hash = hash
	err = a.db.DocumentStore.Create(document)
	if err != nil {
		return err
	}

	documentId = document.Id
	newFile := storage.DocumentPath(document.Id)

	err = storage.CreateDocumentDir(document.Id)
	if err != nil {
		return fmt.Errorf("create directory for doc: %v", err)
	}

	err = storage.MoveFile(tempFileName, newFile)
	if err != nil {
		return fmt.Errorf("rename temp file by document id: %v", err)
	}

	err = a.db.JobStore.AddDocument(document)
	if err != nil {
		return fmt.Errorf("add process steps for new document: %v", err)
	}
	err = a.process.AddDocumentForProcessing(document)
	opOk = true
	return c.JSON(http.StatusOK, responseFromDocument(document))
}

func (a *Api) getEmptyDocument(resp http.ResponseWriter, req *http.Request) {
	doc := &models.Document{}
	respResourceList(resp, responseFromDocument(doc), 1)
}

func (a *Api) downloadDocument(c echo.Context) error {
	// swagger:route GET /api/v1/documents/{id} Documents DownloadDocument
	// Downloads original document
	// Responses:
	//  200: DocumentResponse

	ctx := c.(UserContext)
	var err error
	id := c.Param("id")

	opOk := false
	defer logCrudDocument(ctx.UserId, "download", &opOk, "document: %s", id)
	doc, err := a.db.DocumentStore.GetDocument(ctx.UserId, id)
	if err != nil {
		return err
	}

	filePath := storage.DocumentPath(doc.Id)
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	stat, err := file.Stat()
	size := stat.Size()

	resp := c.Response()
	resp.Header().Set("Content-Type", doc.Mimetype)
	resp.Header().Set("Content-Length", strconv.Itoa(int(size)))
	resp.Header().Set("Cache-Control", "max-age=600")

	_, err = io.Copy(resp, file)
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
	//  200: DocumentResponse

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
	doc, err := a.db.DocumentStore.GetDocument(ctx.UserId, id)
	if err != nil {
		return err
	}

	if dto.Date != 0 {
		doc.Date = time.Unix(dto.Date/1000, 0)
	}

	doc.Name = dto.Name
	doc.Description = dto.Description
	doc.Filename = dto.Filename
	metadata := make([]models.Metadata, len(dto.Metadata))

	for i, v := range dto.Metadata {
		metadata[i] = models.Metadata{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}

	doc.Update()
	doc.Metadata = metadata

	err = a.db.DocumentStore.Update(ctx.UserId, doc)
	if err != nil {
		return err
	}

	err = a.db.MetadataStore.UpdateDocumentKeyValues(ctx.UserId, doc.Id, metadata)
	if err != nil {
		return err
	}

	logrus.Debugf("document updated, force fts update")
	err = a.db.JobStore.ForceProcessing(ctx.UserId, doc.Id, models.ProcessFts)
	if err != nil {
		logrus.Warningf("error marking document for processing (doc %s): %v", doc.Id, err)
	} else {
		err = a.process.AddDocumentForProcessing(doc)
		if err != nil {
			logrus.Warningf("error adding updated document for processing (doc: %s): %v", doc.Id, err)
		}
	}
	opOk = true
	return resourceList(c, responseFromDocument(doc), 1)
}

type documentSortParams struct{}

func (d *documentSortParams) FilterAttributes() []string {
	return []string{"id", "name", "content", "description", "date"}
}

func (d *documentSortParams) SortAttributes() []string {
	return append(d.FilterAttributes(), "created_at", "updated_at")
}

func (d *documentSortParams) SortNoCase() []string { return d.FilterAttributes() }

func (d *documentSortParams) Update() {}

func (a *Api) searchDocuments(userId int, filter *search.DocumentFilter, c echo.Context) error {
	paging, err := bindPaging(c)
	if err != nil {
		return err
	}

	sort, err := getSortParams(c.Request(), &documentSortParams{})
	if err != nil {
		return err
	}

	if len(sort) == 1 {
		filter.Sort = sort[0].Key
		filter.SortMode = strings.ToLower(sort[0].SortOrder())
	}
	if filter.Sort == "id" {
		filter.Sort = ""
	}

	opOk := false
	defer logCrudDocument(userId, "search", &opOk, "metadata: %v, query: %v", filter.Metadata != "", filter.Query != "")

	res, n, err := a.search.SearchDocuments(userId, filter.Query, sort[0], paging)
	if err != nil {
		return err
	}

	docs := make([]*DocumentResponse, len(res))
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
	owns, err := a.db.DocumentStore.UserOwnsDocument(id, ctx.UserId)
	if err != nil {
		return err
	}

	opOk := false
	defer logCrudDocument(ctx.UserId, "schedule processing", &opOk, "document: %s", id)

	if !owns {
		return respForbiddenV2()
	}

	err = a.db.JobStore.ForceProcessing(ctx.UserId, id, models.ProcessRules)
	if err != nil {
		return err
	}

	doc, err := a.db.DocumentStore.GetDocument(ctx.UserId, id)
	if err != nil {
		logrus.Errorf("Get document to process: %v", err)
	} else {
		err = a.process.AddDocumentForProcessing(doc)
		if err != nil {
			logrus.Errorf("schedule document processing: %v", err)
		}
	}
	opOk = true
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

	owns, err := a.db.DocumentStore.UserOwnsDocument(id, ctx.UserId)
	if err != nil {
		return err
	}

	if !owns {
		return respForbiddenV2()
	}

	logrus.Infof("Request user %d removing document %s", ctx.UserId, id)

	err = a.search.DeleteDocument(id, ctx.UserId)
	if err != nil {
		logrus.Errorf("delete document from search index: %v", err)
		return respInternalErrorV2("delete from search index", err)
	}

	_ = process.DeleteDocument(id)
	err = a.db.DocumentStore.DeleteDocument(ctx.UserId, id)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(http.StatusOK, nil)
}

type BulkEditDocumentsRequest struct {
	Documents      []string              `json:"documents" valid:"required"`
	AddMetadata    MetadataUpdateRequest `json:"add_metadata" valid:"-"`
	RemoveMetadata MetadataUpdateRequest `json:"remove_metadata" valid:"-"`
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

	if len(dto.RemoveMetadata.Metadata) == 0 && len(dto.AddMetadata.Metadata) == 0 {
		userErr := errors.ErrAlreadyExists
		userErr.ErrMsg = "no documents modified"
		return userErr
	}
	opOk := false
	defer logCrudDocument(ctx.UserId, "bulk edit", &opOk, "documents: %v, add metadata: %d, remove metadata: %d",
		len(dto.Documents), len(dto.AddMetadata.Metadata), len(dto.RemoveMetadata.Metadata))

	owns, err := a.db.DocumentStore.UserOwnsDocuments(ctx.UserId, dto.Documents)
	if err != nil {
		return err
	}
	if !owns {
		return respForbiddenV2()
	}

	if len(dto.AddMetadata.Metadata) > 0 {
		addMetadata := dto.AddMetadata.toMetadataArray()
		keys := dto.AddMetadata.UniqueKeys()
		ok, err := a.db.MetadataStore.UserHasKeys(ctx.UserId, keys)
		if err != nil {
			return fmt.Errorf("check user owns keys: %v", err)
		}
		if !ok {
			return respForbiddenV2()
		}

		err = a.db.MetadataStore.UpsertDocumentMetadata(ctx.UserId, dto.Documents, addMetadata)
		if err != nil {
			return err
		}
	}
	if len(dto.RemoveMetadata.Metadata) > 0 {
		removeMetadata := dto.RemoveMetadata.toMetadataArray()
		keys := dto.RemoveMetadata.UniqueKeys()
		ok, err := a.db.MetadataStore.UserHasKeys(ctx.UserId, keys)
		if err != nil {
			return fmt.Errorf("check user owns keys: %v", err)
		}
		if !ok {
			return respForbiddenV2()
		}

		err = a.db.MetadataStore.DeleteDocumentsMetadata(ctx.UserId, dto.Documents, removeMetadata)
		if err != nil {
			return err
		}
	}

	// need to reindex
	err = a.db.JobStore.AddDocuments(ctx.UserId, dto.Documents, models.ProcessFts)
	if err != nil {
		if errors.Is(err, errors.ErrAlreadyExists) {
			// already indexing, skip
		} else {
			return err
		}
	}
	a.process.PullDocumentsToProcess()
	opOk = true
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

	owns, err := a.db.DocumentStore.UserOwnsDocument(id, ctx.UserId)
	if err != nil {
		return err
	}

	if !owns {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	data, err := a.db.DocumentStore.GetDocumentHistory(ctx.UserId, id)
	if err != nil {
		return err
	}

	return resourceList(c, data, len(*data))
}
