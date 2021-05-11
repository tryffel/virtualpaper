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
	"bytes"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/process"
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
// swagger:model ForceDocumumentsProcessing
type ForceDocumentProcessingRequest struct {
	UserId     int    `json:"user_id" valid:"-"`
	DocumentId string `json:"document_id" valid:"-"`
	FromStep   string `json:"from_step" valid:"-"`
}

func (a *Api) forceDocumentProcessing(resp http.ResponseWriter, req *http.Request) {
	// swagger:route POST /api/v1/admin/documents/process Admin AdminForceDocumentProcessing
	// Force document processing.
	//
	// Administrator can force re-processing documents.
	// Options:
	// 1. Process all documents in the system. Do not provide user_id or document_id
	// 2. Process documents for a user: provide user_id.
	// 3. Process one document: provide document_id.
	//
	// In addition, step can be configured. Possible steps are:
	// * 1. 'hash' (calculate document hash)
	// * 2. 'thumbnail' (create document thumbnail)
	// * 3. 'content' (extract content with suitable tool)
	// * 4. 'rules' (run metadata-rules)
	// * 5. 'fts' (index document in full-text-search engine)
	//
	// Steps are in order. Supplying e.g. 'content' will result in executing steps 3, 4 and 5.
	// Empty body will result in all documents being processed from step 1.
	// Depending on document content, processing on document takes anywhere from a second to minutes.
	// Consumes:
	// - application/json
	//
	// responses:
	//   200: RespOk
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound

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
	if err != nil {
		respError(resp, err, handler)
	}

	if body.DocumentId != "" {
		doc, err := a.db.DocumentStore.GetDocument(0, body.DocumentId)
		if err != nil {
			logrus.Errorf("Get document to process: %v", err)
		} else {
			err = a.process.AddDocumentForProcessing(doc)
			if err != nil {
				logrus.Errorf("schedule document processing: %v", err)
			}
		}
	}
	respOk(resp, nil)
}

type DocumentProcessStep struct {
	DocumentId string `json:"document_id"`
	Step       string `json:"step"`
}

func (a *Api) getDocumentProcessQueue(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/admin/documents/process Admin AdminGetDocumentProcessQueue
	// Get documents awaiting processing
	//
	// responses:
	//   200: RespDocumentProcessingSteps
	//   401: RespForbidden
	//   500: RespInternalError
	handler := "Api.adminGetProcessQueue"

	queue, n, err := a.db.JobStore.GetPendingProcessing()
	if err != nil {
		respError(resp, err, handler)
		return
	}

	processes := make([]DocumentProcessStep, len(*queue))
	for i, v := range *queue {
		processes[i].DocumentId = v.DocumentId
		processes[i].Step = v.Step.String()
	}

	respResourceList(resp, processes, n)
}

// swagger:response SystemInfo
type SystemInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	GoVersion string `json:"go_version"`

	ImagemagickVersion string `json:"imagemagick_version"`
	TesseractVersion   string `json:"tesseract_version"`
	PopplerInstalled   bool   `json:"poppler_installed"`
	PandocInstalled    bool   `json:"pandoc_installed"`

	NumCpu     int    `json:"number_cpus"`
	ServerLoad string `json:"server_load"`
	Uptime     string `json:"uptime"`

	DocumentsInQueue            int    `json:"documents_queued"`
	DocumentsProcessedToday     int    `json:"documents_processed_today"`
	DocumentsProcessedLastWeek  int    `json:"documents_processed_past_week"`
	DocumentsProcessedLastMonth int    `json:"documents_processed_past_month"`
	DocumentsTotal              int    `json:"documents_total"`
	DocumentsTotalSize          int64  `json:"documents_total_size"`
	DocumentsTotalSizeString    string `json:"documents_total_size_string"`
}

func (a *Api) getSystemInfo(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/admin/systeminfo Admin AdminGetSystemInfo
	// Get system information
	//
	// responses:
	//   200: RespAdminSystemInfo
	//   401: RespForbidden
	//   500: RespInternalError

	info := &SystemInfo{
		Name:               "Virtualpaper",
		Version:            config.Version,
		Commit:             config.Commit,
		NumCpu:             runtime.NumCPU(),
		ImagemagickVersion: process.GetImagickVersion(),
		TesseractVersion:   process.GetTesseractVersion(),
		PopplerInstalled:   process.GetPdfToTextIsInstalled(),
		GoVersion:          runtime.Version(),
		Uptime:             config.UptimeString(),
		PandocInstalled:    process.GetPandocInstalled(),
	}

	stats, err := a.db.StatsStore.GetSystemStats()
	if err != nil {
		respError(resp, err, "getSystemInfo")
		return
	}

	info.DocumentsInQueue = stats.DocumentsInQueue
	info.DocumentsProcessedToday = stats.DocumentsProcessedToday
	info.DocumentsProcessedLastWeek = stats.DocumentsProcessedLastWeek
	info.DocumentsProcessedLastMonth = stats.DocumentsProcessedLastMonth
	info.DocumentsTotal = stats.DocumentsTotal
	info.DocumentsTotalSize = stats.DocumentsTotalSize
	info.DocumentsTotalSizeString = models.GetPrettySize(stats.DocumentsTotalSize)

	stdout := &bytes.Buffer{}
	cmd := exec.Command("uptime")
	cmd.Stdout = stdout
	err = cmd.Run()

	if err != nil {
		logrus.Warningf("exec 'uptime': %v", err)
	} else {
		text := stdout.String()
		text = strings.Trim(text, " \n")
		splits := strings.Split(text, " ")
		if len(splits) != 13 {
			logrus.Warningf("invalid 'uptime' result: %v", splits)
		} else {
			load := strings.Join(splits[10:], " ")
			info.ServerLoad = load
		}
	}

	respOk(resp, info)
}
