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
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

type documentResponse struct {
	Id          int               `json:"id"`
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

func responseFromDocument(doc *models.Document) *documentResponse {
	resp := &documentResponse{
		Id:          doc.Id,
		Name:        doc.Name,
		Filename:    doc.Filename,
		Content:     doc.Content,
		Description: doc.Description,
		CreatedAt:   doc.CreatedAt.Unix() * 1000,
		UpdatedAt:   doc.UpdatedAt.Unix() * 1000,
		Date:        doc.Date.Unix() * 1000,
		PreviewUrl:  fmt.Sprintf("%s/api/v1/documents/%d/preview", config.C.Api.PublicUrl, doc.Id),
		DownloadUrl: fmt.Sprintf("%s/api/v1/documents/%d/download", config.C.Api.PublicUrl, doc.Id),
		Mimetype:    doc.Mimetype,
		Type:        doc.GetType(),
		Size:        doc.Size,
		PrettySize:  doc.GetSize(),
		Metadata:    doc.Metadata,
		Tags:        doc.Tags,
	}
	return resp
}

type documentUpdateRequest struct {
	Name        string `json:"name" valid:"-"`
	Description string `json:"description" valid:"-"`
	Filename    string `json:"filename" valid:"ascii"`
	Date        int64  `json:"date" valid:"-"`
}

func (a *Api) getDocuments(resp http.ResponseWriter, req *http.Request) {
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

	docs, count, err := a.db.DocumentStore.GetDocuments(user, paging, true)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		respInternalError(resp)
		return
	}
	respDocs := make([]*documentResponse, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}

	respResourceList(resp, respDocs, count)
}

func (a *Api) getDocument(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getDocument"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	idStr := mux.Vars(req)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respBadRequest(resp, "id not integer", nil)
		return
	}

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
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	idStr := mux.Vars(req)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respBadRequest(resp, "id not integer", nil)
		return
	}

	content, err := a.db.DocumentStore.GetContent(user, id)
	respOk(resp, &content)

}

func (a *Api) getDocumentLogs(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getDocumentLogs"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id, err := getParamId(req)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
		return
	}

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
	respOk(resp, job)
}

func (a *Api) getDocumentPreview(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getDocumentPreview"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	id, err := getParamId(req)
	if err != nil {
		respBadRequest(resp, err.Error(), nil)
		return
	}

	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	file, err := os.OpenFile(path.Join(config.C.Processing.PreviewsDir, doc.Hash+".png"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// TODO: mark document missing thumbnail
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
			respError(resp, storage.ErrInternalError, handler)
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
	header.Set("Content-Disposition", "attachment; filename="+doc.Hash+".png")
	resp.Header().Set("Cache-Control", "max-age=600")

	defer file.Close()

	_, err = io.Copy(resp, file)
	if err != nil {
		respError(resp, err, handler)
	} else {
		respOk(resp, nil)
	}

}

func (a *Api) uploadFile(resp http.ResponseWriter, req *http.Request) {
	handler := "api.uploadFile"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"), handler)
	}
	var err error
	err = req.ParseMultipartForm(1024 * 1024 * 500)
	reader, header, err := req.FormFile("file")
	if err != nil {
		userError := storage.ErrInvalid
		userError.ErrMsg = err.Error()
		userError.Err = err
		respError(resp, userError, handler)
		return
	}

	mimetype := header.Header.Get("Content-Type")
	defer reader.Close()

	hash := config.RandomString(10)

	document := &models.Document{
		Id:       0,
		UserId:   userId,
		Name:     header.Filename,
		Content:  "",
		Filename: header.Filename,
		Hash:     hash,
		Mimetype: mimetype,
		Size:     header.Size,
		Date:     time.Now(),
	}

	file, err := os.OpenFile(path.Join(config.C.Processing.DocumentsDir, hash), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		respError(resp, fmt.Errorf("open new file for saving upload: %v", err), handler)
		return
	}
	defer file.Close()

	n, err := file.ReadFrom(reader)
	if err != nil {
		respError(resp, fmt.Errorf("write uploaded file to disk: %v", err), handler)
		return
	}

	if n != header.Size {
		logrus.Warningf("did not fully read file: %d, got: %d", header.Size, n)
	}
	defer file.Close()

	err = a.db.DocumentStore.Create(document)
	if err != nil {
		respError(resp, err, handler)
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
	respOk(resp, responseFromDocument(doc))
}

func (a *Api) downloadDocument(resp http.ResponseWriter, req *http.Request) {
	handler := "download document"
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"), handler)
	}
	var err error

	idStr := mux.Vars(req)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respBadRequest(resp, "id not integer", nil)
		return
	}

	doc, err := a.db.DocumentStore.GetDocument(userId, id)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	file, err := os.Open(path.Join(config.C.Processing.DocumentsDir, doc.Hash))
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
		respError(resp, err, handler)
		return
	}

}

func (a *Api) updateDocument(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.updateDocument"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}
	idStr := mux.Vars(req)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respBadRequest(resp, "id not integer", nil)
		return
	}

	dto := &documentUpdateRequest{}
	err = unMarshalBody(req, dto)
	if err != nil {
		respError(resp, err, handler)
		return
	}

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

	doc.Update()

	err = a.db.DocumentStore.Update(doc)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	err = a.search.IndexDocuments(&[]models.Document{*doc}, user)
	if err != nil {
		respError(resp, err, handler)
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

	res, n, err := a.search.SearchDocuments(userId, filter, paging)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	docs := make([]*documentResponse, len(res))
	for i, v := range res {
		docs[i] = responseFromDocument(v)
	}
	respResourceList(resp, docs, n)
}
