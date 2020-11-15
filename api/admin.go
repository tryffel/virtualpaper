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

func (a *Api) authorizeAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		admin, err := userIsAdmin(r)
		if err != nil {
			respError(w, err, "api.authorizeAdmin")
			return
		}
		if !admin {
			if logrus.IsLevelEnabled(logrus.DebugLevel) {
				user, ok := getUser(r)
				if !ok {
					logrus.Debug("user (unknown) is not admin, refuse to serve")
				} else {
					logrus.Debugf("user %d is not admin, refuse to serve", user.Id)
				}
			}
			respUnauthorized(w)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ForceDocumentsProcessingRequest describes request to force processing of documents.
type ForceDocumentProcessingRequest struct {
	UserId     int    `json:"user_id" valid:"-"`
	DocumentId string `json:"document_id" valid:"-"`
	FromStep   string `json:"from_step" valid:"-"`
}

func (a *Api) forceDocumentProcessing(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/admin/documents/process Admin AdminForceDocumentProcessing
	// Force document processing
	//
	// responses:
	//   200:
	handler := "Api.getDocuments"
	body := &ForceDocumentProcessingRequest{}
	err := unMarshalBody(req, body)
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
	case "rules":
		step = models.ProcessRules
	case "fts":
		step = models.ProcessFts
	default:
		respBadRequest(resp, "no such step", nil)
		return
	}

	err = a.db.JobStore.ForceProcessing(body.UserId, body.DocumentId, step)
	if err == nil {
		respOk(resp, nil)
		return
	}

	respError(resp, err, handler)
}

type documentProcessStep struct {
	DocumentId string `json:"document_id"`
	Step       string `json:"step"`
}

func (a *Api) getDocumentProcessQueue(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/admin/documents/process Admin AdminGetDocumentProcessQueue
	// Get documents awaiting processing
	//
	// responses:
	//   200: DocumentResponse
	handler := "Api.adminGetProcessQueue"

	queue, n, err := a.db.JobStore.GetPendingProcessing()
	if err != nil {
		respError(resp, err, handler)
		return
	}

	processes := make([]documentProcessStep, len(*queue))
	for i, v := range *queue {
		processes[i].DocumentId = v.DocumentId
		processes[i].Step = v.Step.String()
	}

	respResourceList(resp, processes, n)
}
