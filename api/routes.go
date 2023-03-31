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
	"golang.org/x/time/rate"
	"time"
	"tryffel.net/go/virtualpaper/config"
)

func (api *Api) addRoutesV2() {
	api.publicRouter = api.echo.Group("")
	api.apiRouter = api.publicRouter.Group("/api")
	api.privateRouter = api.apiRouter.Group("/v1", api.authorizeUserV2())
	api.adminRouter = api.privateRouter.Group("/admin", api.AuthorizeAdminV2())

	//a.oldPrivateRouter.HandleFunc("/tags/{id}", a.getTag).Methods(http.MethodGet)
	//a.oldPrivateRouter.HandleFunc("/tags", a.createTag).Methods(http.MethodPost)
	//a.oldPrivateRouter.HandleFunc("/tags/create", a.createTag).Methods(http.MethodPost)

	api.publicRouter.StaticFS("/", static())
	api.publicRouter.GET("/api/v1/swagger.json", serverSwaggerDoc)
	api.publicRouter.GET("/api/v1/version", api.getVersionV2)

	// allow one auth operation per minute for past 15 minutes, with burst of 15 requests.

	authGroup := api.apiRouter.Group("/v1/auth")
	if !config.C.Api.AuthRatelimitDisabled {
		authRateLimiter := newRateLimiter(rate.Every(time.Second*60), 15, time.Minute*15)
		authGroup = api.apiRouter.Group("/v1/auth", authRateLimiter)
	}

	authGroup.POST("/login", api.LoginV2)
	api.privateRouter.POST("/auth/logout", api.Logout)
	api.privateRouter.POST("/auth/confirm", api.ConfirmAuthentication)
	authGroup.POST("/reset-password", api.ResetPassword)
	authGroup.POST("/forgot-password", api.CreateResetPasswordToken)

	api.privateRouter.GET("/filetypes", api.getSupportedFileTypes)
	api.privateRouter.GET("/admin/systeminfo", api.getSystemInfo)

	api.privateRouter.GET("/documents/stats", api.getUserDocumentStatistics)
	api.privateRouter.POST("/documents", api.uploadFile)
	api.privateRouter.GET("/documents", api.getDocuments).Name = "get-documents"
	api.privateRouter.GET("/documents/deleted", api.getDeletedDocuments).Name = "get-deleted-documents"
	api.privateRouter.GET("/documents/:id", api.getDocument).Name = "get-document"
	api.privateRouter.PUT("/documents/:id", api.updateDocument)
	api.privateRouter.DELETE("/documents/:id", api.deleteDocument)
	api.privateRouter.POST("/documents/deleted/:id/restore", api.restoreDeletedDocument)
	api.privateRouter.GET("/documents/:id/show", api.getDocument).Name = "get-document"
	api.privateRouter.GET("/documents/:id/preview", api.getDocumentPreview)
	api.privateRouter.GET("/documents/:id/content", api.getDocumentContent)
	api.privateRouter.GET("/documents/:id/download", api.downloadDocument)
	api.privateRouter.GET("/documents/:id/linked-documents", api.getLinkedDocuments)
	api.privateRouter.POST("/documents/:id/metadata", api.updateDocumentMetadata)
	api.privateRouter.POST("/documents/:id/process", api.requestDocumentProcessing)
	api.privateRouter.PUT("/documents/:id/linked-documents", api.updateLinkedDocuments)
	api.privateRouter.GET("/documents/:id/history", api.getDocumentHistory)
	api.privateRouter.GET("/documents/:id/jobs", api.getDocumentLogs)

	api.privateRouter.POST("/documents/bulkEdit", api.bulkEditDocuments)

	api.privateRouter.POST("/documents/search/suggest", api.searchSuggestions).Name = "search-suggest"

	api.privateRouter.GET("/jobs", api.GetJob)
	api.privateRouter.GET("/tags", api.getTags)

	api.privateRouter.GET("/metadata/keys", api.getMetadataKeys)
	api.privateRouter.POST("/metadata/keys", api.addMetadataKey)
	api.privateRouter.PUT("/metadata/keys/:id", api.updateMetadataKey)
	api.privateRouter.GET("/metadata/keys/:id", api.getMetadataKey)
	api.privateRouter.GET("/metadata/keys/:id/values", api.getMetadataKeyValues)
	api.privateRouter.POST("/metadata/keys/:id/values", api.addMetadataValue)
	api.privateRouter.DELETE("/metadata/keys/:id", api.deleteMetadataKey)
	api.privateRouter.PUT("/metadata/keys/:keyId/values/:valueId", api.updateMetadataValue)
	api.privateRouter.DELETE("/metadata/keys/:keyId/values/:valueId", api.deleteMetadataValue)

	api.privateRouter.GET("/processing/rules", api.getUserRules)
	api.privateRouter.POST("/processing/rules", api.addUserRule)
	api.privateRouter.GET("/processing/rules/:id", api.getUserRule)
	api.privateRouter.PUT("/processing/rules/:id", api.updateUserRule)
	api.privateRouter.DELETE("/processing/rules/:id", api.deleteUserRule)
	api.privateRouter.PUT("/processing/rules/:id/test", api.testRule)

	api.privateRouter.GET("/preferences/user", api.getUserPreferences).Name = "get-user-preferences"
	api.privateRouter.PUT("/preferences/user", api.updateUserPreferences)

	api.adminRouter.GET("/documents/process", api.getDocumentProcessQueue)
	api.adminRouter.POST("/documents/process", api.forceDocumentProcessing)
	api.adminRouter.POST("/documents/deleted/:id/restore", api.adminRestoreDeletedDocument)

	api.adminRouter.GET("/users", api.adminGetUsers)
	api.adminRouter.POST("/users", api.adminAddUser, api.ConfirmAuthorizedToken())
	api.adminRouter.GET("/users/:id", api.adminGetUser)
	api.adminRouter.PUT("/users/:id", api.adminUpdateUser, api.ConfirmAuthorizedToken())
}
