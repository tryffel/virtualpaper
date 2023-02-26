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
	"github.com/labstack/echo/v4"
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
	YearlyStats []models.UserDocumentYearStat `json:"yearly_stats"`
	// total number of metadata keys
	// Example: 4
	NumMetadataKeys int `json:"num_metadata_keys"`
	// total number of metadata values
	// Example: 14
	NumMetadataValues int `json:"num_metadata_values"`
	// array of last updated document ids
	// Example: [abcd]
	LastDocumentsUpdated []string `json:"last_documents_updated"`
	LastDocumentsAdded   []string `json:"last_documents_added"`
	LastDocumentsViewed  []string `json:"last_documents_viewed"`

	Indexing bool `json:"indexing"`
}

func docStatsToUserStats(stats *models.UserDocumentStatistics) *UserDocumentStatistics {
	uds := &UserDocumentStatistics{
		UserId:               stats.UserId,
		NumDocuments:         stats.NumDocuments,
		YearlyStats:          stats.YearlyStats,
		NumMetadataKeys:      stats.NumMetadataKeys,
		NumMetadataValues:    stats.NumMetadataValues,
		LastDocumentsUpdated: stats.LastDocumentsUpdated,
		LastDocumentsAdded:   stats.LastDocumentsAdded,
		LastDocumentsViewed:  stats.LastDocumentsViewed,
	}

	if uds.LastDocumentsUpdated == nil {
		uds.LastDocumentsUpdated = []string{}
	}
	if uds.LastDocumentsAdded == nil {
		uds.LastDocumentsAdded = []string{}
	}
	if uds.LastDocumentsViewed == nil {
		uds.LastDocumentsViewed = []string{}
	}

	if uds.YearlyStats == nil {
		uds.YearlyStats = []models.UserDocumentYearStat{}
	}
	return uds
}

func (a *Api) getUserDocumentStatistics(c echo.Context) error {
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

	ctx := c.(UserContext)
	stats, err := a.db.StatsStore.GetUserDocumentStats(ctx.UserId)
	if err != nil {
		return err
	} else {
		stats.UserId = ctx.UserId
		statsDto := docStatsToUserStats(stats)
		searchStats, err := a.search.GetUserIndexStatus(ctx.UserId)
		if err != nil {
			logrus.Warningf("get search engine indexing status: %v", err)
		} else {
			statsDto.Indexing = searchStats.Indexing
		}
		return c.JSON(http.StatusOK, statsDto)
	}
}
