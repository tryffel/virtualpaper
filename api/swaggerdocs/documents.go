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

package swaggerdocs

import (
	"gopkg.in/h2non/gentleman.v2/plugins/multipart"
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

// DocumentsResponse contains array of documents
// swagger:response Document
type documentResponse struct {
	// in:body
	Body []aggregates.Document
}

// DocumentExistsResponse contains existing document's id and error message.
// swagger:response DocumentExistsResponse
type documentExistsResponse struct {
	// in:body
	Body api.DocumentExistsResponse
}

// Upload file
// swagger:parameters UploadFile ReqUploadFile
type UploadFileRequest struct {
	// in:body
	Body multipart.FormData
}

// Bulk edit documents
// swagger:parameters BulkEditDocuments ReqBulkEditDocuments
type BulkEditDocuments struct {
	// in:body
	Body api.BulkEditDocumentsRequest
}

// swagger:parameters GetDocuments ReqDocumentFilter
type documentFilter struct {
	// Json filter containing max two keys: q and metadata.
	// Q is full-text-search query.
	// Metadata is a metadata filter.
	// E.g. 'class:book AND (author:"agatha christie" OR author:"doyle")'
	// Filter is json-formatted and must be url-safe.
	// example: '{"q":"my search", "metadata":"class:book"}'
	// in: query
	// required: false
	Filter string `json:"filter"`
	// Order which order results in, either: 'DESC' or 'ASC'.
	Order string `json:"order"`
	// Sort field to sort results.
	Sort string `json:"sort"`
	// Page number
	// required: false
	Page int `json:"page"`
	// Page size.
	// required: false
	PerPage int `json:"perPage"`

	/*
			// Full-text-search query
			//
			Query string `json:"q"`
			// Metadata filter
			// example: "class:book AND (author:"agatha christie" OR author:"doyle")"
			//
			MetadataFilter string `json:"metadata"`
		}

	*/
}
