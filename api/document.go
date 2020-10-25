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
	"tryffel.net/go/virtualpaper/storage"
)

type documentResponse struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	Date        int64  `json:"date"`
	PreviewUrl  string `json:"preview_url"`
	DownloadUrl string `json:"download_url"`
	Mimetype    string `json:"mimetype"`
	Type        string `json:"type"`
	Size        int64  `json:"size"`
	PrettySize  string `json:"pretty_size"`
	Status      string `json:"status"`
}

func responseFromDocument(doc *models.Document) *documentResponse {
	resp := &documentResponse{
		Id:          doc.Id,
		Name:        doc.Name,
		Filename:    doc.Filename,
		Content:     doc.Content,
		CreatedAt:   doc.CreatedAt.Unix() * 1000,
		UpdatedAt:   doc.UpdatedAt.Unix() * 1000,
		Date:        doc.Date.Unix() * 1000,
		PreviewUrl:  fmt.Sprintf("%s/api/v1/documents/%d/preview", config.C.Api.PublicUrl, doc.Id),
		DownloadUrl: fmt.Sprintf("%s/api/v1/documents/%d/download", config.C.Api.PublicUrl, doc.Id),
		Mimetype:    doc.Mimetype,
		Type:        doc.GetType(),
		Size:        doc.Size,
		PrettySize:  doc.GetSize(),
	}
	return resp
}

type documentUpdateRequest struct {
	Name     string `json:"name" valid:"required"`
	Filename string `json:"filename" valid:"ascii,required"`
	Date     int64  `json:"date" valid:"required"`
}

func (a *Api) getDocuments(resp http.ResponseWriter, req *http.Request) {
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
		respError(resp, err)
		return
	}

	status, err := a.db.JobStore.GetDocumentStatus(doc.Id)
	if err != nil {
		respError(resp, err)
		return
	}

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
		respError(resp, err)
		return
	}

	if !owns {
		respUnauthorized(resp)
		return
	}

	job, err := a.db.JobStore.GetByDocument(id)
	if err != nil {
		logrus.Errorf("get document jobs: %v", err)
		respError(resp, err)
		return
	}
	respOk(resp, job)
}

func (a *Api) getDocumentPreview(resp http.ResponseWriter, req *http.Request) {
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
		respError(resp, err)
		return
	}

	file, err := os.OpenFile(path.Join(config.C.Processing.PreviewsDir, doc.Hash+".png"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// TODO: mark document missing thumbnail
			logrus.Warningf("document %d does not have thumbnail", id)
			respError(resp, storage.ErrRecordNotFound)
			return
		}
		respError(resp, err)
		return
	}
	stat, err := file.Stat()
	if err != nil {
		respError(resp, fmt.Errorf("get file preview stat: %v", err))
		return
	}

	header := resp.Header()
	header.Set("Content-Type", "image/png")
	header.Set("Content-Length", strconv.Itoa(int(stat.Size())))
	header.Set("Content-Disposition", "attachment; filename="+doc.Hash+".png")

	defer file.Close()

	_, err = io.Copy(resp, file)
	if err != nil {
		respError(resp, err)
	} else {
		respOk(resp, nil)
	}

}

func (a *Api) uploadFile(resp http.ResponseWriter, req *http.Request) {
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"))
	}
	var err error
	err = req.ParseMultipartForm(1024 * 1024 * 500)
	reader, header, err := req.FormFile("file")
	if err != nil {
		respError(resp, fmt.Errorf("parse multipart"))
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
	}

	file, err := os.OpenFile(path.Join(config.C.Processing.DocumentsDir, hash), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		respError(resp, fmt.Errorf("open new file for saving upload: %v", err))
		return
	}
	defer file.Close()

	n, err := file.ReadFrom(reader)
	if err != nil {
		respError(resp, fmt.Errorf("write uploaded file to disk: %v", err))
		return
	}

	if n != header.Size {
		logrus.Warningf("did not fully read file: %d, got: %d", header.Size, n)
	}
	defer file.Close()

	err = a.db.DocumentStore.Create(document)
	if err != nil {
		respError(resp, fmt.Errorf("new document: %v", err))
		return
	}

	err = a.db.JobStore.AddDocument(document)
	if err != nil {
		respError(resp, fmt.Errorf("add process steps for new document: %v", err))
		return
	}
	respOk(resp, responseFromDocument(document))
	return
}

func (a *Api) getEmptyDocument(resp http.ResponseWriter, req *http.Request) {
	doc := &models.Document{}
	respOk(resp, responseFromDocument(doc))
}

func (a *Api) downloadDocument(resp http.ResponseWriter, req *http.Request) {
	userId, ok := getUserId(req)
	if !ok {
		respError(resp, errors.New("no userId found"))
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
		respError(resp, err)
		return
	}

	file, err := os.Open(path.Join(config.C.Processing.DocumentsDir, doc.Hash))
	if err != nil {
		respError(resp, err)
		return
	}

	defer file.Close()

	stat, err := file.Stat()
	size := stat.Size()

	resp.Header().Set("Content-Type", doc.Mimetype)
	resp.Header().Set("Content-Length", strconv.Itoa(int(size)))

	_, err = io.Copy(resp, file)
	if err != nil {
		respError(resp, err)
		return
	}

}

func (a *Api) updateDocument(resp http.ResponseWriter, req *http.Request) {
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
		respError(resp, err)
		return
	}

	doc, err := a.db.DocumentStore.GetDocument(user, id)
	if err != nil {
		respError(resp, err)
		return
	}

	doc.Name = dto.Name
	doc.Filename = dto.Filename
	doc.Date = time.Unix(dto.Date/1000, 0)

	doc.Update()

	err = a.db.DocumentStore.Update(doc)
	if err != nil {
		respError(resp, err)
	} else {
		respResourceList(resp, responseFromDocument(doc), 1)
	}
}
