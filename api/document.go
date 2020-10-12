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

type documentResponse struct {
	Id        int
	Name      string
	Filename  string
	Content   string
	CreatedAt PrettyTime
	UpdatedAt PrettyTime
}

func responseFromDocument(doc *models.Document) *documentResponse {
	resp := &documentResponse{
		Id:        doc.Id,
		Name:      doc.Name,
		Filename:  doc.Filename,
		Content:   doc.Content,
		CreatedAt: PrettyTime(doc.CreatedAt),
		UpdatedAt: PrettyTime(doc.UpdatedAt),
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

	docs, err := a.db.DocumentStore.GetDocuments(user, 100)
	if err != nil {
		logrus.Errorf("get documents: %v", err)
		respInternalError(resp)
		return
	}
	respDocs := make([]*documentResponse, len(*docs))

	for i, v := range *docs {
		respDocs[i] = responseFromDocument(&v)
	}

	body := map[string]interface{}{"documents": respDocs}
	respOk(resp, body)
}
