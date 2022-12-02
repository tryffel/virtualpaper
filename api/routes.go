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
	"net/http"
	"tryffel.net/go/virtualpaper/config"
)

func (a *Api) addRoutes() {
	if len(config.C.Api.CorsHosts) > 0 {
		a.baseRouter.Use(a.corsHeader)
	}

	a.apiRouter.Use(LoggingMiddleware)
	a.baseRouter.Handle("/", staticServer()).Methods(http.MethodGet)
	a.baseRouter.HandleFunc("/api/v1/auth/login", a.login).Methods(http.MethodPost)
	a.baseRouter.HandleFunc("/api/v1/version", a.getVersion).Methods(http.MethodGet)
	a.baseRouter.HandleFunc("/api/v1/swagger.json", serverSwaggerDoc).Methods(http.MethodGet)

	a.privateRouter.Use(a.authorizeUser)
	a.privateRouter.HandleFunc("/documents", a.getDocuments).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/undefined", a.getEmptyDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/show", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id:[a-zA-Z0-9-]{30,40}}", a.getDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}", a.deleteDocument).Methods(http.MethodDelete)
	a.privateRouter.HandleFunc("/documents/{id}", a.updateDocument).Methods(http.MethodPut)
	a.privateRouter.HandleFunc("/documents/{id}/preview", a.getDocumentPreview).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/jobs", a.getDocumentLogs).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/create", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/undefined", a.uploadFile).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/create", a.getEmptyDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/bulkEdit", a.bulkEditDocuments).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/{id}/content", a.getDocumentContent).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/download", a.downloadDocument).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/metadata", a.updateDocumentMetadata).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/{id}/process", a.requestDocumentProcessing).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/documents/{id}/history", a.getDocumentHistory).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/linked-documents", a.getLinkedDocuments).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/documents/{id}/linked-documents", a.updateLinkedDocuments).Methods(http.MethodPut)

	a.privateRouter.HandleFunc("/documents/search/suggest", a.searchSuggestions).Methods(http.MethodPost)

	a.privateRouter.HandleFunc("/jobs", a.GetJob).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/tags", a.getTags).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/tags/{id}", a.getTag).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/tags", a.createTag).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/tags/create", a.createTag).Methods(http.MethodPost)

	a.privateRouter.HandleFunc("/metadata/keys", a.getMetadataKeys).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys", a.addMetadataKey).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/metadata/keys/{id}", a.updateMetadataKey).Methods(http.MethodPut)
	a.privateRouter.HandleFunc("/metadata/keys/{id}", a.getMetadataKey).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys/{id}", a.deleteMetadataKey).Methods(http.MethodDelete)
	a.privateRouter.HandleFunc("/metadata/keys/{id}/values", a.getMetadataKeyValues).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/metadata/keys/{id}/values", a.addMetadataValue).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/metadata/keys/{key_id}/values/{value_id}", a.updateMetadataValue).Methods(http.MethodPut)
	a.privateRouter.HandleFunc("/metadata/keys/{key_id}/values/{value_id}", a.deleteMetadataValue).Methods(http.MethodDelete)

	a.privateRouter.HandleFunc("/documents/stats", a.getUserDocumentStatistics).Methods(http.MethodGet)

	a.privateRouter.HandleFunc("/processing/rules", a.addUserRule).Methods(http.MethodPost)
	a.privateRouter.HandleFunc("/processing/rules/{id}", a.updateUserRule).Methods(http.MethodPut)
	a.privateRouter.HandleFunc("/processing/rules/{id}", a.deleteUserRule).Methods(http.MethodDelete)
	a.privateRouter.HandleFunc("/processing/rules", a.getUserRules).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/processing/rules/{id}", a.getUserRule).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/processing/rules/{id}/test", a.testRule).Methods(http.MethodPut)

	a.privateRouter.HandleFunc("/preferences/user", a.getUserPreferences).Methods(http.MethodGet)
	a.privateRouter.HandleFunc("/preferences/user", a.updateUserPreferences).Methods(http.MethodPut)

	a.privateRouter.HandleFunc("/filetypes", a.getSupportedFileTypes).Methods(http.MethodGet)

	a.adminRouter.Use(a.authorizeUser, a.authorizeAdmin)
	a.adminRouter.HandleFunc("/documents/process", a.forceDocumentProcessing).Methods(http.MethodPost)
	a.adminRouter.HandleFunc("/documents/process", a.getDocumentProcessQueue).Methods(http.MethodGet)
	a.adminRouter.HandleFunc("/users", a.getUsers).Methods(http.MethodGet)

	// allow non-admins access to system info
	a.privateRouter.HandleFunc("/admin/systeminfo", a.getSystemInfo).Methods(http.MethodGet)
}
