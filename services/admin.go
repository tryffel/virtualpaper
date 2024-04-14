package services

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/models/aggregates"
	"tryffel.net/go/virtualpaper/services/process"
	"tryffel.net/go/virtualpaper/services/search"
	"tryffel.net/go/virtualpaper/storage"
	"tryffel.net/go/virtualpaper/util/logger"
)

type AdminService struct {
	db      *storage.Database
	process *process.Manager
	search  *search.Engine
}

func NewAdminService(db *storage.Database, manager *process.Manager, search *search.Engine) *AdminService {
	return &AdminService{
		db:      db,
		process: manager,
		search:  search,
	}
}

func (service *AdminService) GetDocumentProcessQueue(ctx context.Context) (*[]models.ProcessItem, int, error) {
	return service.db.JobStore.GetPendingProcessing()
}

func (service *AdminService) GetSystemInfo(ctx context.Context) (*aggregates.SystemInfo, error) {
	info := &aggregates.SystemInfo{
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
		ProcessingStatus:   service.process.ProcessingStatus(),
		ProcessingEnabled:  !config.C.Processing.Disabled,
		CronJobsEnabled:    !config.C.CronJobs.Disabled,
	}

	stats, err := service.db.StatsStore.GetSystemStats()
	if err != nil {
		return nil, err
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

	engineStatus, err := service.search.GetStatus()
	if err != nil {
		logrus.Errorf("get search engine status: %v", err)
	} else {
		info.SearchEngineStatus = *engineStatus
	}
	return info, nil
}

func (service *AdminService) GetUsers() (*[]models.UserInfo, error) {
	info, err := service.db.UserStore.GetUsersInfo()
	if err != nil {
		return nil, err
	}

	//searchStatus, _, err := service.search.GetUserIndicesStatus()
	//if err != nil {
	//return nil, err
	//}

	/*
		for i, v := range *info {
			//indexStatus := searchStatus[v.UserId]
			if indexStatus != nil {
				(*info)[i].Indexing = indexStatus.Indexing
				(*info)[i].TotalDocumentsIndexed = indexStatus.NumDocuments
			}
		}

	*/
	return info, nil
}

func (service *AdminService) GetUser(ctx context.Context, id int) (*models.UserInfo, error) {
	userInfo, err := service.db.UserStore.GetUser(id)
	if err != nil {
		return nil, err
	}

	searchStatus, err := service.search.GetIndexStatus()
	if err != nil {
		return nil, err
	}

	info := &models.UserInfo{
		UserId:        userInfo.Id,
		UserName:      userInfo.Name,
		Email:         userInfo.Email,
		IsActive:      userInfo.IsActive,
		UpdatedAt:     userInfo.UpdatedAt,
		CreatedAt:     userInfo.CreatedAt,
		DocumentCount: 0,
		DocumentsSize: 0,
		IsAdmin:       userInfo.IsAdmin,
		LastSeen:      time.Time{},
		Indexing:      searchStatus.Indexing,
	}
	return info, nil
}

func (service *AdminService) UpdateUser(ctx context.Context, adminUser int, updated *models.User) (*models.UserInfo, error) {
	user, err := service.db.UserStore.GetUser(updated.Id)
	if err != nil {
		return nil, err
	}

	dataChanged := false
	if user.IsActive != updated.IsActive {
		if updated.IsActive {
			logger.Context(ctx).Infof("Activate user %d by admin user %d", user.Id, adminUser)
		} else {
			logger.Context(ctx).Infof("Deactivate user %d by admin user %d", user.Id, adminUser)
		}
		user.IsActive = updated.IsActive
		dataChanged = true
	}
	if user.IsAdmin != updated.IsAdmin {
		if updated.IsAdmin {
			logger.Context(ctx).Infof("Add user %d to administrators by admin user %d", user.Id, adminUser)
		} else {
			logger.Context(ctx).Infof("Remove user %d from administrators by admin user %d", user.Id, adminUser)
		}
		user.IsAdmin = updated.IsAdmin
		dataChanged = true
	}
	if user.Email != updated.Email {
		logger.Context(ctx).Infof("Change user's %d email by admin user %d", user.Id, adminUser)
		user.Email = updated.Email
		dataChanged = true
	}
	if updated.Password != "" {
		logger.Context(ctx).Infof("Change user's %d password by admin user %d", user.Id, adminUser)
		err = user.SetPassword(updated.Password)
		if err != nil {
			return nil, fmt.Errorf("set user's password: %v", err)
		}
		dataChanged = true
	}
	if dataChanged {
		user.Update()
		err = service.db.UserStore.Update(user)
		if err == nil {
			info := &models.UserInfo{
				UserId:        user.Id,
				UserName:      user.Name,
				Email:         user.Email,
				IsActive:      user.IsActive,
				UpdatedAt:     user.UpdatedAt,
				CreatedAt:     user.CreatedAt,
				DocumentCount: 0,
				DocumentsSize: 0,
				IsAdmin:       user.IsAdmin,
				LastSeen:      time.Time{},
				Indexing:      false,
			}
			return info, nil
		}
		return nil, err
	}
	return &models.UserInfo{
		UserId:        user.Id,
		UserName:      user.Name,
		Email:         user.Email,
		IsActive:      user.IsActive,
		UpdatedAt:     user.UpdatedAt,
		CreatedAt:     user.CreatedAt,
		DocumentCount: 0,
		DocumentsSize: 0,
		IsAdmin:       user.IsAdmin,
		LastSeen:      time.Time{},
		Indexing:      false,
	}, nil
}

type NewUser struct {
	Name     string
	Email    string
	Admin    bool
	Active   bool
	Password string
}

func (service *AdminService) CreateUser(ctx context.Context, adminUser int, newUser NewUser) (*models.UserInfo, error) {
	user := &models.User{
		Timestamp: models.Timestamp{},
		Id:        0,
		Name:      newUser.Name,
		Email:     newUser.Email,
		IsAdmin:   newUser.Admin,
		IsActive:  newUser.Active,
	}
	err := user.SetPassword(newUser.Password)
	if err != nil {
		return nil, fmt.Errorf("set password: %v", err)
	}
	user.CreatedAt = time.Now()
	user.Update()

	err = service.db.UserStore.AddUser(user)
	if err != nil {
		return nil, err
	}

	if user.IsAdmin {
		logger.Context(ctx).Infof("admin user %d created new user %d with admin privileges", adminUser, user.Id)
	}
	createdUser, err := service.db.UserStore.GetUserByName(newUser.Name)
	if err != nil {
		return nil, err
	}
	info := &models.UserInfo{
		UserId:        createdUser.Id,
		UserName:      user.Name,
		Email:         user.Email,
		IsActive:      user.IsActive,
		UpdatedAt:     user.UpdatedAt,
		CreatedAt:     user.CreatedAt,
		DocumentCount: 0,
		DocumentsSize: 0,
		IsAdmin:       user.IsAdmin,
		LastSeen:      time.Time{},
		Indexing:      false,
	}
	return info, nil
}

func (service *AdminService) RestoreDeletedDocument(ctx context.Context, adminUserId int, id string) error {
	document, err := service.db.DocumentStore.GetDocument(service.db, id)
	if err != nil {
		return err
	}
	if !document.DeletedAt.Valid {
		return errors.ErrRecordNotFound
	}

	err = service.db.DocumentStore.MarkDocumentNonDeleted(service.db, adminUserId, id)
	if err != nil {
		return err
	}

	doc, err := service.db.DocumentStore.GetDocument(service.db, id)
	err = service.search.IndexDocuments(&[]models.Document{*doc}, doc.UserId)
	if err != nil {
		logrus.Errorf("delete document from search index: %v", err)
	}
	return nil
}

func (service *AdminService) ForceProcessingByUser(ctx context.Context, userId int, steps []models.ProcessStep) error {
	return service.db.JobStore.ForceProcessingByUser(userId, steps)
}

func (service *AdminService) ForceProcessingByDocumentId(ctx context.Context, docId string, steps []models.ProcessStep) error {
	doc, err := service.db.DocumentStore.GetDocument(service.db, docId)
	if err != nil {
		return err
	}
	err = service.db.JobStore.ForceProcessingDocument(service.db, doc.Id, steps)
	if err != nil {
		return err
	}
	err = service.process.AddDocumentForProcessing(doc.Id)
	if err != nil {
		return err
	}
	service.process.PullDocumentsToProcess()
	return nil
}
