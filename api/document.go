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

// DocumentUpdateRequest
// swagger:model DocumentUpdateRequestBody
type DocumentUpdateRequest struct {
	Name        string            `json:"name" valid:"-"`
	Description string            `json:"description" valid:"-"`
	Filename    string            `json:"filename" valid:"-"`
	Date        int64             `json:"date" valid:"-"`
	Metadata    []MetadataRequest `json:"metadata" valid:"-"`
}

func (a *Api) getDocuments(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents Documents GetDocuments
	// Get documents
	//
	// responses:
	//   200: DocumentResponse
	handler := "Api.getDocuments"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	query, err := getDocumentFilter(req)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if query != nil {
		a.searchDocuments(user, query, resp, req)
		return
	}

	paging, err := getPaging(req)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
	}

	sort, err := getSortParams(req, &models.Document{})
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if len(sort) == 0 {
		sort = append(sort, storage.SortKey{})
	}

	docs, count, err := a.db.DocumentStore.GetDocuments(user, paging, sort[0], true)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		respInternalError(resp)
		return
	}
	respDocs := make([]*DocumentResponse, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}

	respResourceList(resp, respDocs, count)
}

func (a *Api) getDocument(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/{id} Documents GetDocument
	// Get document
	// responses:
	//   200: DocumentResponse
	handler := "Api.getDocument"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)
	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	status, err := a.db.JobStore.GetDocumentStatus(doc.Id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	metadata, err := a.db.MetadataStore.GetDocumentMetadata(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	doc.Metadata = *metadata

	tags, err := a.db.MetadataStore.GetDocumentTags(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	doc.Tags = *tags

	respDoc := responseFromDocument(doc)
	respDoc.Status = status
	respOk(resp, respDoc)
}

func (a *Api) getDocumentContent(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/{id}/content Documents GetDocumentContent
	// Get full document parsed content
	// responses:
	//   200: DocumentResponse
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)

	content, err := a.db.DocumentStore.GetContent(user, id)
	if err != nil {
		respError(resp, err, "api.getDocumentContent")
		return
	}
	respOk(resp, &content)
}

func (a *Api) getDocumentLogs(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/{id}/jobs Documents GetDocumentJobs
	// Get processing job history related to document
	// responses:
	//   200: DocumentResponse

	handler := "Api.getDocumentLogs"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)
	owns, err := a.db.DocumentStore.UserOwnsDocument(id, user)
	if err != nil {
		logrus.Errorf("Get document ownserhip: %v", err)
		respError(resp, err, handler)
		return
	}

	if !owns {
		respUnauthorized(resp)
		return
	}

	job, err := a.db.JobStore.GetByDocument(id)
	if err != nil {
		logrus.Errorf("get document jobs: %v", err)
		respError(resp, err, handler)
		return
	}
	respResourceList(resp, job, len(*job))
}

func (a *Api) getDocumentPreview(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/{id}/preview Documents GetDocumentPreview
	// Get document preview, a small png image of first page of document.
	// responses:
	//   200: DocumentResponse
	handler := "Api.getDocumentPreview"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)
	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	filePath := storage.PreviewPath(doc.Id)
	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			/*
				logrus.Warningf("document %d does not have thumbnail, scheduling", id)
				process := &models.ProcessItem{
					DocumentId: doc.Id,
					Document:   nil,
					Step:       models.ProcessThumbnail,
					CreatedAt:  time.Now(),
				}

				err = a.db.JobStore.CreateProcessItem(process)
				if err != nil {
					if errors.Is(err, storage.ErrAlreadyExists) {
					} else {
						logrus.Error("add process step for missing thumbnail (doc %c): %v", doc.Id, err)
					}
				}

				err = a.process.AddDocumentForProcessing(doc)
				if err != nil {
					if errors.Is(err, storage.ErrAlreadyExists) {
						// process exists already
					} else {
						logrus.Error("schedule document thumbnail (doc %c): %v", doc.Id, err)
					}
				}

			*/
			respError(resp, errors.ErrInternalError, handler)
			return
		}
		respError(resp, err, handler)
		return
	}
	stat, err := file.Stat()
	if err != nil {
		respError(resp, fmt.Errorf("get file preview stat: %v", err), handler)
		return
	}

	header := resp.Header()
	header.Set("Content-Type", "image/png")
	header.Set("Content-Length", strconv.Itoa(int(stat.Size())))
	header.Set("Content-Disposition", "attachment; filename="+doc.Id+".png")
	resp.Header().Set("Cache-Control", "max-age=600")

	defer file.Close()

	_, err = io.Copy(resp, file)
	if err != nil {
		logrus.Errorf("send file over http: %v", err)
	}
}

func (a *Api) uploadFile(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/documents Documents UploadFile
	// Upload new document file. New document already contains id, name, filename and timestamps.
	// Otherwise document is not processed yet and lacks other fields.
	// Responses:
	//  200: DocumentResponse
	handler := "api.uploadFile"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"), handler)
	}
	var err error
	err = req.ParseMultipartForm(1024 * 1024 * 500)
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid form: %v", err)
		userError.Err = err
		respError(resp, userError, handler)
		return
	}
	reader, header, err := req.FormFile("file")
	if err != nil {
		userError := errors.ErrInvalid
		userError.ErrMsg = fmt.Sprintf("invalid file: %v", err)
		userError.Err = err
		respError(resp, userError, handler)
		return
	}

	mimetype := header.Header.Get("Content-Type")

	if mimetype == "application/octet-stream" {
		mimetype = "text/plain"
	}

	defer reader.Close()

	tempHash := config.RandomString(10)

	document := &models.Document{
		Id:       "",
		UserId:   userId,
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
		respError(resp, e, handler)
		req.Body.Close()
		return
	}

	tempFileName := storage.TempFilePath(tempHash)
	inputFile, err := os.OpenFile(tempFileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		respError(resp, fmt.Errorf("open new file for saving upload: %v", err), handler)
		return
	}
	n, err := inputFile.ReadFrom(reader)
	if err != nil {
		respError(resp, fmt.Errorf("write uploaded file to disk: %v", err), handler)
		return
	}

	if n != header.Size {
		logrus.Warningf("did not fully read file: %d, got: %d", header.Size, n)
	}

	err = inputFile.Close()
	if err != nil {
		respError(resp, fmt.Errorf("close file: %v", err), handler)
		return
	}

	hash, err := process.GetHash(tempFileName)
	if err != nil {
		respError(resp, fmt.Errorf("get hash for temp file: %v", err), handler)
		return
	}

	existingDoc, err := a.db.DocumentStore.GetByHash(hash)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
		} else {
			respError(resp, fmt.Errorf("get existing document by hash: %v", err), handler)
			return
		}
	}

	if existingDoc != nil {
		if existingDoc.Id != "" {
			_ = respJson(resp, fmt.Sprintf(`{id: %s}`, existingDoc.Id), http.StatusNotModified)
			err := os.Remove(tempFileName)
			if err != nil {
				logrus.Errorf("remove duplicated temp file: %v", err)
			}
			return
		}
	}

	document.Hash = hash
	err = a.db.DocumentStore.Create(document)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	newFile := storage.DocumentPath(document.Id)

	err = storage.CreateDocumentDir(document.Id)
	if err != nil {
		logrus.Errorf("create directory for doc: %v", err)
		return
	}

	err = storage.MoveFile(tempFileName, newFile)
	if err != nil {
		logrus.Errorf("rename temp file by document id: %v", err)
		return
	}

	err = a.db.JobStore.AddDocument(document)
	if err != nil {
		respError(resp, fmt.Errorf("add process steps for new document: %v", err), handler)
		return
	}
	err = a.process.AddDocumentForProcessing(document)
	if err != nil {
		respError(resp, err, handler)
	} else {
		respOk(resp, responseFromDocument(document))
	}
	return
}

func (a *Api) getEmptyDocument(resp http.ResponseWriter, req *http.Request) {
	doc := &models.Document{}
	respResourceList(resp, responseFromDocument(doc), 1)
}

func (a *Api) downloadDocument(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/{id} Documents DownloadDocument
	// Downloads original document
	// Responses:
	//  200: DocumentResponse
	handler := "download document"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"), handler)
	}
	var err error

	id := getParamId(req)
	doc, err := a.db.DocumentStore.GetDocument(userId, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	filePath := storage.DocumentPath(doc.Id)
	file, err := os.Open(filePath)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	defer file.Close()

	stat, err := file.Stat()
	size := stat.Size()

	resp.Header().Set("Content-Type", doc.Mimetype)
	resp.Header().Set("Content-Length", strconv.Itoa(int(size)))
	resp.Header().Set("Cache-Control", "max-age=600")

	_, err = io.Copy(resp, file)
	if err != nil {
		logrus.Errorf("send file over http: %v", err)
	}
}

func (a *Api) updateDocument(resp http.ResponseWriter, req *http.Request) {
	// swagger:route PUT /api/v1/documents/{id} Documents UpdateDocument
	// Updates document
	// Responses:
	//  200: DocumentResponse

	handler := "Api.updateDocument"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	id := getParamId(req)
	dto := &DocumentUpdateRequest{}
	err := unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	dto.Filename = govalidator.SafeFileName(dto.Filename)
	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if dto.Date != 0 {
		doc.Date = time.Unix(dto.Date/1000, 0)
	}

	doc.Name = dto.Name
	doc.Description = dto.Description
	doc.Filename = dto.Filename
	metadata := make([]*models.Metadata, len(dto.Metadata))

	for i, v := range dto.Metadata {
		metadata[i] = &models.Metadata{
			KeyId:   v.KeyId,
			ValueId: v.ValueId,
		}
	}

	doc.Update()

	err = a.db.DocumentStore.Update(doc)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	err = a.db.MetadataStore.UpdateDocumentKeyValues(user, doc.Id, metadata)
	if err != nil {
		respError(resp, err, handler)
	}

	logrus.Debugf("document updated, force fts update")
	err = a.db.JobStore.ForceProcessing(user, doc.Id, models.ProcessFts)
	if err != nil {
		logrus.Warningf("error marking document for processing (doc %s): %v", doc.Id, err)
	} else {
		err = a.process.AddDocumentForProcessing(doc)
		if err != nil {
			logrus.Warningf("error adding updated document for processing (doc: %s): %v", doc.Id, err)
		}
	}
	respResourceList(resp, responseFromDocument(doc), 1)
}

func (a *Api) searchDocuments(userId int, filter *search.DocumentFilter, resp http.ResponseWriter, req *http.Request) {
	handler := "api.searchDocuments"

	paging, err := getPaging(req)
	if err != nil {
		logrus.Warningf("invalid paging: %v", err)
		paging.Limit = 100
		paging.Offset = 0
	}

	sort, err := getSortParams(req, &models.Document{})
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if len(sort) == 1 {
		filter.Sort = sort[0].Key
		filter.SortMode = strings.ToLower(sort[0].SortOrder())
	}

	res, n, err := a.search.SearchDocuments(userId, filter, paging)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	docs := make([]*DocumentResponse, len(res))
	for i, v := range res {
		docs[i] = responseFromDocument(v)
	}
	respResourceList(resp, docs, n)
}

func (a *Api) requestDocumentProcessing(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/location Documents RequestProcessing
	// Request document re-processing
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	handler := "Api.requestDocumentProcessing"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)

	owns, err := a.db.DocumentStore.UserOwnsDocument(id, user)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if !owns {
		respForbidden(resp)
		return
	}

	err = a.db.JobStore.ForceProcessing(user, id, models.ProcessRules)
	if err != nil {
		respError(resp, err, handler)
	}

	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		logrus.Errorf("Get document to process: %v", err)
	} else {
		err = a.process.AddDocumentForProcessing(doc)
		if err != nil {
			logrus.Errorf("schedule document processing: %v", err)
		}
	}
	respOk(resp, nil)
}

func (a *Api) deleteDocument(resp http.ResponseWriter, req *http.Request) {
	// swagger:route DELETE /api/v1/documents/:id Documents DeleteDocument
	// Delete document
	// Responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError

	handler := "Api.deleteDocument"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id := getParamId(req)
	owns, err := a.db.DocumentStore.UserOwnsDocument(id, user)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	if !owns {
		respForbidden(resp)
		return
	}

	logrus.Infof("Request user %d removing document %s", user, id)

	err = a.search.DeleteDocument(id, user)
	if err != nil {
		logrus.Errorf("delete document from search index: %v", err)
		respInternalError(resp)
	}

	process.DeleteDocument(id)
	err = a.db.DocumentStore.DeleteDocument(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}
	respOk(resp, nil)
}
