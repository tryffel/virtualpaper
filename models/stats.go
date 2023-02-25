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

package models

// UserDocumentStatistics contains various about user's
// documents.
type UserDocumentStatistics struct {
	UserId       int `json:"id"`
	NumDocuments int `json:"num_documents"`
	YearlyStats  []struct {
		Year         int `json:"year" db:"year"`
		NumDocuments int `json:"num_documents" db:"count"`
	} `json:"yearly_stats"`
	NumMetadataKeys      int      `json:"num_metadata_keys"`
	NumMetadataValues    int      `json:"num_metadata_values"`
	LastDocumentsUpdated []string `json:"last_documents_updated"`
	LastDocumentsAdded   []string `json:"last_documents_added"`
	LastDocumentsViewed  []string `json:"last_documents_viewed"`
}

type SystemStatistics struct {
	DocumentsInQueue            int    `json:"documents_queued" db:"documents_queued"`
	DocumentsProcessedToday     int    `json:"documents_processed_today" db:"documents_processed_today"`
	DocumentsProcessedLastWeek  int    `json:"documents_processed_past_week" db:"documents_processed_past_week"`
	DocumentsProcessedLastMonth int    `json:"documents_processed_past_month" db:"documents_processed_past_month"`
	DocumentsTotal              int    `json:"documents_total" db:"documents_total"`
	DocumentsTotalSize          int64  `json:"documents_total_size" db:"documents_size"`
	DocumentsTotalSizeString    string `json:"documents_total_size_string"`
	ServerLoad                  string `json:"server_load"`
}
