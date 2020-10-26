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

// ForceDocumentsProcessingRequest describes request to force processing of documents.
type ForceDocumentProcessingRequest struct {
	UserId     int    `json:"user_id" valid:"-"`
	DocumentId int    `json:"document_id" valid:"-"`
	FromStep   string `json:"from_step" valid:"-"`
}

func (a *Api) forceDocumentProcessing(resp http.ResponseWriter, req *http.Request) {
	handler := "Api.getDocuments"
	userId, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	user, err := a.db.UserStore.GetUser(userId)
	if err != nil {
		respError(resp, err, handler)
	}

	if !user.IsAdmin {
		respUnauthorized(resp)
		return
	}

	body := &ForceDocumentProcessingRequest{}
	err = unMarshalBody(req, body)
	if err != nil {
		respError(resp, err, handler)
		return
	}

	step := models.ProcessFts
	switch body.FromStep {
	case "hash":
		step = models.ProcessHash
	case "thumbnail":
		step = models.ProcessThumbnail
	case "content":
		step = models.ProcessParseContent
	case "fts":
		step = models.ProcessFts
	default:
		step = -1
	}

	err = a.db.JobStore.ForceProcessing(body.UserId, body.DocumentId, step)
	if err == nil {
		respOk(resp, nil)
		return
	}

	respError(resp, err, handler)
}
