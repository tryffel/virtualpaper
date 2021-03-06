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

type UserDocumentStatistics struct {
	// user id
	UserId int `json:"id"`
	// total number of documents
	// Example: 53
	NumDocuments int `json:"num_documents"`
	// per-year statistics
	YearlyStats []struct {
		// year
		// Example: 2020
		Year int `json:"year" db:"year"`
		// number of documents
		// Example: 49
		NumDocuments int `json:"num_documents" db:"count"`
	} `json:"yearly_stats"`
	// total number of metadata keys
	// Example: 4
	NumMetadataKeys int `json:"num_metadata_keys"`
	// total number of metadata values
	// Example: 14
	NumMetadataValues int `json:"num_metadata_values"`
	// array of last updated document ids
	// Example: [abcd]
	LastDocumentsUpdated []string `json:"last_documents_updated"`
}

func docStatsToUserStats(stats *models.UserDocumentStatistics) *UserDocumentStatistics {
	uds := &UserDocumentStatistics{
		UserId:               stats.UserId,
		NumDocuments:         stats.NumDocuments,
		YearlyStats:          stats.YearlyStats,
		NumMetadataKeys:      stats.NumMetadataKeys,
		NumMetadataValues:    stats.NumMetadataValues,
		LastDocumentsUpdated: stats.LastDocumentsUpdated,
	}

	if uds.YearlyStats == nil {
		uds.YearlyStats = []struct {
			Year         int `json:"year" db:"year"`
			NumDocuments int `json:"num_documents" db:"count"`
		}{}
	}
	return uds
}

func (a *Api) getUserDocumentStatistics(resp http.ResponseWriter, req *http.Request) {
	// swagger:route GET /api/v1/documents/stats Documents GetUserDocumentStatistics
	// Get document statistics
	//
	// responses:
	//   200: RespDocumentStatistics
	//   304: RespNotModified
	//   400: RespBadRequest
	//   401: RespForbidden
	//   403: RespNotFound
	//   500: RespInternalError
	handler := "Api.getUserDocumentStatistics"
	user, ok := getUserId(req)
	if !ok {
		logrus.Errorf("no user in context")
		respInternalError(resp)
		return
	}

	stats, err := a.db.StatsStore.GetUserDocumentStats(user)
	if err != nil {
		respError(resp, err, handler)
	} else {
		stats.UserId = user
		respOk(resp, docStatsToUserStats(stats))
	}
}
