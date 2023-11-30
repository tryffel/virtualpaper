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
	"tryffel.net/go/virtualpaper/services"
	"tryffel.net/go/virtualpaper/services/process"
)

func (a *Api) AuthorizeAdminV2() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, ok := c.(UserContext)
			if !ok {
				c.Logger().Error("no UserContext found")
				return echo.ErrInternalServerError
			}
			if !ctx.Admin {
				return echo.ErrUnauthorized
			}

			return next(c)
		}
	}
}

// ForceDocumentsProcessingRequest describes request to force processing of documents.
// swagger:model ForceDocumumentsProcessing
type ForceDocumentProcessingRequest struct {
	UserId     int    `json:"user_id" valid:"-"`
	DocumentId string `json:"document_id" valid:"-"`
	FromStep   string `json:"from_step" valid:"process_step~invalid process step"`
}

func (a *Api) forceDocumentProcessing(c echo.Context) error {
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

	body := &ForceDocumentProcessingRequest{}
	err := unMarshalBody(c.Request(), body)
	if err != nil {
		return err
	}

	step := models.ProcessFts
	switch body.FromStep {
	case "hash":
		step = models.ProcessHash
	case "thumbnail":
		step = models.ProcessThumbnail
	case "content":
		step = models.ProcessParseContent
	case "detect-language":
		step = models.ProcessDetectLanguage
	case "rules":
		step = models.ProcessRules
	case "fts":
		step = models.ProcessFts
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step")
	}
	steps := append(process.RequiredProcessingSteps(step), step)

	if body.UserId != 0 {
		err = a.db.JobStore.ForceProcessingByUser(body.UserId, steps)
	} else {
		err = a.db.JobStore.ForceProcessingDocument(body.DocumentId, steps)
		if err != nil {
			return err
		}
	}

	if body.DocumentId != "" {
		doc, err := a.db.DocumentStore.GetDocument(body.DocumentId)
		if err != nil {
			logrus.Errorf("Get document to process: %v", err)
		} else {
			err = a.process.AddDocumentForProcessing(doc.Id)
			if err != nil {
				logrus.Errorf("schedule document processing: %v", err)
			}
		}
	} else {
		a.process.PullDocumentsToProcess()
	}
	return c.String(http.StatusOK, "")
}

type DocumentProcessStep struct {
	DocumentId string `json:"id"`
	Step       string `json:"step"`
}

func (a *Api) getDocumentProcessQueue(c echo.Context) error {
	// swagger:route GET /api/v1/admin/documents/process Admin AdminGetDocumentProcessQueue
	// Get documents awaiting processing
	//
	// responses:
	//   200: RespDocumentProcessingSteps
	//   401: RespForbidden
	//   500: RespInternalError

	queue, n, err := a.adminService.GetDocumentProcessQueue(getContext(c))
	if err != nil {
		return err
	}

	processes := make([]DocumentProcessStep, len(*queue))
	for i, v := range *queue {
		processes[i].DocumentId = v.DocumentId
		processes[i].Step = v.Action.String()
	}

	return resourceList(c, processes, n)
}

func (a *Api) getSystemInfo(c echo.Context) error {
	// swagger:route GET /api/v1/admin/systeminfo Admin AdminGetSystemInfo
	// Get system information
	//
	// responses:
	//   200: RespAdminSystemInfo
	//   401: RespForbidden
	//   500: RespInternalError

	info, err := a.adminService.GetSystemInfo(getContext(c))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, info)
}

func (a *Api) adminGetUsers(c echo.Context) error {
	// swagger:route GET /api/v1/admin/users Admin AdminGetUsers
	// Get detailed users info.
	//
	// responses:
	//   200: RespUserInfo
	ctx := c.(UserContext)
	opOk := false
	defer func() {
		logCrudAdminUsers(ctx.UserId, "list", &opOk, "get users")
	}()

	info, err := a.adminService.GetUsers()
	if err != nil {
		return err
	}
	opOk = true
	return resourceList(c, info, len(*info))
}

func (a *Api) adminGetUser(c echo.Context) error {
	// swagger:route GET /api/v1/admin/user Admin AdminGetUser
	// Get detailed user info
	//
	// responses:
	//   200: RespUserInfo
	ctx := c.(UserContext)
	userId, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	opOk := false
	defer func() {
		logCrudAdminUsers(ctx.UserId, "get", &opOk, "get user %d", userId)
	}()

	info, err := a.adminService.GetUser(getContext(ctx), userId)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, info)
}

type AdminUpdateUserRequest struct {
	Email         string `json:"email" valid:"optional,email"`
	Password      string `json:"password" valid:"optional"`
	Active        bool   `json:"is_active" valid:"optional"`
	Administrator bool   `json:"is_admin" valid:"optional"`
}

func (a *Api) adminUpdateUser(c echo.Context) error {
	// swagger:route PUT /api/v1/admin/users/:id Admin AdminUpdateUser
	// UpdateJob user
	//
	// responses:
	//   200: RespUserInfo

	ctx := c.(UserContext)
	request := &AdminUpdateUserRequest{}
	err := unMarshalBody(c.Request(), request)
	if err != nil {
		return err
	}
	if request.Password != "" {
		if err = ValidatePassword(request.Password); err != nil {
			return err
		}
	}

	userId, err := bindPathIdInt(c)
	if err != nil {
		return err
	}

	opOk := false
	defer func() {
		logCrudAdminUsers(ctx.UserId, "update", &opOk, "update user, user_id: %d", userId)
	}()

	user := &models.User{
		Timestamp: models.Timestamp{},
		Id:        userId,
		Name:      "",
		Password:  request.Password,
		Email:     request.Email,
		IsAdmin:   request.Administrator,
		IsActive:  request.Active,
	}

	info, err := a.adminService.UpdateUser(getContext(c), ctx.UserId, user)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, info)
}

type AdminAddUserRequest struct {
	UserName      string `json:"user_name" valid:"username"`
	Email         string `json:"email" valid:"email,optional"`
	Password      string `json:"password" valid:"required"`
	Active        bool   `json:"is_active" valid:"optional"`
	Administrator bool   `json:"is_admin" valid:"optional"`
}

func (a *Api) adminAddUser(c echo.Context) error {
	// swagger:route POST /api/v1/admin/users/ Admin AdminAddUser
	// Add new user
	//
	// responses:
	//   200: RespUserInfo

	ctx := c.(UserContext)
	request := &AdminAddUserRequest{}
	err := unMarshalBody(c.Request(), request)
	if err != nil {
		return err
	}

	if err = ValidatePassword(request.Password); err != nil {
		return err
	}

	opOk := false
	userId := -1
	defer func() {
		logCrudAdminUsers(ctx.UserId, "create", &opOk, "add user %d", userId)
	}()
	newUser := services.NewUser{
		Name:     request.UserName,
		Email:    request.Email,
		Admin:    request.Administrator,
		Active:   request.Active,
		Password: request.Password,
	}
	info, err := a.adminService.CreateUser(getContext(c), ctx.UserId, newUser)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, info)
}

func (a *Api) adminRestoreDeletedDocument(c echo.Context) error {
	// swagger:route PUT /api/v1/admin/documents/trashbin/:id/restore Admin AdminRestoreDeletedDocument
	// Restore deleted document
	//
	// responses:
	//   200: RespUserInfo

	ctx := c.(UserContext)
	docId := bindPathId(c)

	opOk := false
	defer func() {
		logCrudAdminUsers(ctx.UserId, "restore deleted document", &opOk, "restore document %s", docId)
	}()
	err := a.adminService.RestoreDeletedDocument(getContext(ctx), ctx.UserId, docId)
	if err != nil {
		return err
	}
	opOk = true
	return c.JSON(200, nil)
}
