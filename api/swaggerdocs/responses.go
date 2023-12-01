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
	"tryffel.net/go/virtualpaper/api"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
)

// Request ok
// swagger:response RespOk
type RespOk struct{}

// Content already exists and has not been modified
// swagger:response RespNotModified
type RespNotModified struct{}

// request validation failed
// swagger:response RespBadRequest
type RespBadRequest struct {
	// in:body
	Body struct {
		Error string `json:"error"`
	}
}

// action forbidden
// swagger:response RespForbidden
type RespForbidden struct{}

// resource not found
// swagger:response RespNotFound
type RespNotFound struct{}

// internal error occured and server was unable to complete operation
// swagger:response RespInternalError
type RespInternalError struct{}

// User preferences
// swagger:response RespUserPreferences
type UserPreferences struct {
	// in:body
	Body api.UserPreferences
}

// Documents and processing steps pending
// swagger:response RespDocumentProcessingSteps
type DocumentProcessingResp struct {
	// in:body
	Body []api.DocumentProcessStep
}

// Document / usage statistics
// swagger:response RespDocumentStatistics
type UserDocumentStatistics struct {
	// in:body
	Body aggregates.UserDocumentStatistics
}

// System information
// swagger:response RespAdminSystemInfo
type AdminSystemInfo struct {
	// in:body
	Body aggregates.SystemInfo
}

// Force processing documents
// swagger:parameters AdminForceDocumentProcessing ReqForceDocumentsProcessing
type AdminForceDocumentsProcessing struct {
	// in:body
	Body api.ForceDocumentProcessingRequest
}

// User info
// swagger:response RespUserInfo
type RespUserInfo struct {
	// in:body
	Body []models.UserInfo
}

// Supported file types
// swagger:response RespFileTypes
type RespFileTypes struct {
	// in:body
	Body []api.MimeTypesSupportedResponse
}

// Server version
// swagger:response RespVersion
type RespVersion struct {
	//in:body
	Body api.VersionResponse
}
