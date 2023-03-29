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
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/process"
	"tryffel.net/go/virtualpaper/search"
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
	FromStep   string `json:"from_step" valid:"-"`
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
	case "rules":
		step = models.ProcessRules
	case "fts":
		step = models.ProcessFts
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid step")
	}

	err = a.db.JobStore.ForceProcessing(body.UserId, body.DocumentId, step)
	if err != nil {
		return err
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
	queue, n, err := a.db.JobStore.GetPendingProcessing()
	if err != nil {
		return err
	}

	processes := make([]DocumentProcessStep, len(*queue))
	for i, v := range *queue {
		processes[i].DocumentId = v.DocumentId
		processes[i].Step = v.Step.String()
	}

	return resourceList(c, processes, n)
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

	ProcessingStatus   []process.QueueStatus `json:"processing_queue"`
	SearchEngineStatus search.EngineStatus   `json:"search_engine_status"`

	ProcessingEnabled bool `json:"processing_enabled"`
	CronJobsEnabled   bool `json:"cronjobs_enabled"`
}

func (a *Api) getSystemInfo(c echo.Context) error {
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
		ProcessingStatus:   a.process.ProcessingStatus(),
		ProcessingEnabled:  !config.C.Processing.Disabled,
		CronJobsEnabled:    !config.C.CronJobs.Disabled,
	}

	stats, err := a.db.StatsStore.GetSystemStats()
	if err != nil {
		return err
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

	engineStatus, err := a.search.GetStatus()
	if err != nil {
		logrus.Errorf("get search engine status: %v", err)
	} else {
		info.SearchEngineStatus = *engineStatus
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

	info, err := a.db.UserStore.GetUsersInfo()
	if err != nil {
		return err
	}

	searchStatus, _, err := a.search.GetUserIndicesStatus()
	if err != nil {
		return err
	}

	for i, v := range *info {
		indexStatus := searchStatus[v.UserId]
		if indexStatus != nil {
			(*info)[i].Indexing = indexStatus.Indexing
			(*info)[i].TotalDocumentsIndexed = indexStatus.NumDocuments
		}
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

	userInfo, err := a.db.UserStore.GetUser(userId)
	if err != nil {
		return err
	}

	searchStatus, err := a.search.GetUserIndexStatus(userId)
	if err != nil {
		return err
	}

	info := models.UserInfo{
		UserId:                userInfo.Id,
		UserName:              userInfo.Name,
		Email:                 userInfo.Email,
		IsActive:              userInfo.IsActive,
		UpdatedAt:             userInfo.UpdatedAt,
		CreatedAt:             userInfo.CreatedAt,
		DocumentCount:         0,
		DocumentsSize:         0,
		IsAdmin:               userInfo.IsAdmin,
		LastSeen:              time.Time{},
		Indexing:              searchStatus.Indexing,
		TotalDocumentsIndexed: searchStatus.NumDocuments,
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
	// Update user
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

	user, err := a.db.UserStore.GetUser(userId)
	if err != nil {
		return fmt.Errorf("get user %d: %v", userId, err)
	}

	dataChanged := false
	if user.IsActive != request.Active {
		if request.Active {
			logrus.Infof("Activate user %d by admin user %d", user.Id, ctx.UserId)
		} else {
			logrus.Infof("Deactivate user %d by admin user %d", user.Id, ctx.UserId)
		}
		user.IsActive = request.Active
		dataChanged = true
	}
	if user.IsAdmin != request.Administrator {
		if request.Administrator {
			logrus.Infof("Add user %d to administrators by admin user %d", user.Id, ctx.UserId)
		} else {
			logrus.Infof("Remove user %d from administrators by admin user %d", user.Id, ctx.UserId)
		}
		user.IsAdmin = request.Administrator
		dataChanged = true
	}
	if user.Email != request.Email {
		logrus.Infof("Change user's %d email by admin user %d", user.Id, ctx.UserId)
		user.Email = request.Email
		dataChanged = true
	}
	if request.Password != "" {
		logrus.Infof("Change user's %d password by admin user %d", user.Id, ctx.UserId)
		err = user.SetPassword(request.Password)
		if err != nil {
			return fmt.Errorf("set user's password: %v", err)
		}
		dataChanged = true
	}
	if dataChanged {
		user.Update()
		err = a.db.UserStore.Update(user)
		if err == nil {
			info := models.UserInfo{
				UserId:                user.Id,
				UserName:              user.Name,
				Email:                 user.Email,
				IsActive:              user.IsActive,
				UpdatedAt:             user.UpdatedAt,
				CreatedAt:             user.CreatedAt,
				DocumentCount:         0,
				DocumentsSize:         0,
				IsAdmin:               user.IsAdmin,
				LastSeen:              time.Time{},
				Indexing:              false,
				TotalDocumentsIndexed: 0,
			}
			opOk = true
			return c.JSON(200, info)
		}
	} else {
		opOk = true
		return c.JSON(http.StatusNotModified, user)
	}
	return err
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

	user := &models.User{
		Timestamp: models.Timestamp{},
		Id:        0,
		Name:      request.UserName,
		Email:     request.Email,
		IsAdmin:   request.Administrator,
		IsActive:  request.Active,
	}
	err = user.SetPassword(request.Password)
	if err != nil {
		return fmt.Errorf("set password: %v", err)
	}
	user.CreatedAt = time.Now()
	user.Update()

	err = a.db.UserStore.AddUser(user)
	if err != nil {
		return err
	}

	if user.IsAdmin {
		logrus.Infof("admin user %d created new user %d with admin privileges", ctx.UserId, user.Id)
	}

	err = a.search.AddUserIndex(user.Id)
	info := &models.UserInfo{
		UserId:                user.Id,
		UserName:              user.Name,
		Email:                 user.Email,
		IsActive:              user.IsActive,
		UpdatedAt:             user.UpdatedAt,
		CreatedAt:             user.CreatedAt,
		DocumentCount:         0,
		DocumentsSize:         0,
		IsAdmin:               user.IsAdmin,
		LastSeen:              time.Time{},
		Indexing:              false,
		TotalDocumentsIndexed: 0,
	}
	opOk = true
	return c.JSON(200, info)
}
