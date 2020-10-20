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
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
)

type documentResponse struct {
	Id        int `json:"id"`
	Name      string
	Filename  string
	Content   string
	CreatedAt PrettyTime
	UpdatedAt PrettyTime
	Url       string
	Mimetype  string
}

func responseFromDocument(doc *models.Document) *documentResponse {
	resp := &documentResponse{
		Id:        doc.Id,
		Name:      doc.Name,
		Filename:  doc.Filename,
		Content:   doc.Content,
		CreatedAt: PrettyTime(doc.CreatedAt),
		UpdatedAt: PrettyTime(doc.UpdatedAt),
		Url:       fmt.Sprintf("/api/v1/documents/%d/preview", doc.Id),
		Mimetype:  doc.Mimetype,
	}
	return resp
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

	docs, count, err := a.db.DocumentStore.GetDocuments(user, paging)
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

	respOk(resp, responseFromDocument(doc))
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
