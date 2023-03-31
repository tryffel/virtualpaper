package process

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/storage"
)

type CronJobs struct {
	c  *cron.Cron
	db *storage.Database

	removeExpiredPasswordPresets cron.EntryID
	removeExpiredAuthTokens      cron.EntryID
	cleanupDocumenTrashbins      cron.EntryID
}

func NewCron(db *storage.Database) (*CronJobs, error) {
	cj := &CronJobs{
		c:  cron.New(),
		db: db,
	}
	var err error
	cj.removeExpiredPasswordPresets, err = cj.c.AddFunc("*/15 * * * *", cj.JobRemoveExpiredPasswordResets)
	if err != nil {
		return cj, fmt.Errorf("create removeExpiredPasswordPresets job: %v", err)
	}
	cj.removeExpiredAuthTokens, err = cj.c.AddFunc("*/15 * * * *", cj.JobRemoveExpiredAuthTokens)
	if err != nil {
		return cj, fmt.Errorf("create removeExpiredAuthTokens job: %v", err)
	}
	cj.cleanupDocumenTrashbins, err = cj.c.AddFunc("*/1 * * * *", cj.JobCleanupDocumenTrashbins)
	if err != nil {
		return cj, fmt.Errorf("create removeExpiredAuthTokens job: %v", err)
	}
	return cj, nil
}

func (c *CronJobs) Start() {
	if config.C.CronJobs.Disabled {
		logrus.Warningf("cronjobs disabled, refuse to start jobs")
		return
	}

	logrus.Info("starting cron jobs")
	c.c.Start()
}

func (c *CronJobs) Stop() {
	logrus.Info("stopping cron jobs")
	ctx := c.c.Stop()

	timeout := time.Second * 10
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		logrus.Warning("cron jobs killed")
		return
	case <-ctx.Done():
		logrus.Infof("cron jobs stopped")
		timer.Stop()
		return
	}
}

func (c *CronJobs) recover() {
	if r := recover(); r != nil {
		logrus.Errorf("panic in cron: %v", r)
	}
}

func logCronOp(action string, success bool) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"module":    "cron",
		"operation": action,
		"success":   success,
	})
}

func (c *CronJobs) JobRemoveExpiredPasswordResets() {
	defer c.recover()
	action := "remove expired password reset tokens"
	count, err := c.db.UserStore.DeleteExpiredPasswordResetTokens()
	if err != nil {
		logCronOp(action, false).Error(err)
	} else {
		logCronOp(action, true).Debugf("deleted %d tokens", count)
	}
}

func (c *CronJobs) JobRemoveExpiredAuthTokens() {
	defer c.recover()
	action := "remove expired auth tokens"
	count, err := c.db.AuthStore.DeleteExpiredAuthTokens()
	if err != nil {
		logCronOp(action, false).Error(err)
	} else {
		logCronOp(action, true).Debugf("deleted %d tokens", count)
	}
}

func (c *CronJobs) JobCleanupDocumenTrashbins() {
	defer c.recover()
	action := "remove documents marked as deleted"
	if config.C.CronJobs.DocumentsTrashbinDuration.Milliseconds() == 0 {
		logrus.Debugf("deleted documents cleanup period is set to 0, skip removing deleted documents")
		return
	}
	timestamp := time.Now().Add(-config.C.CronJobs.DocumentsTrashbinDuration)

	logrus.Debugf("delete documents marked deleted_as before '%s'", timestamp.String())
	documentsToDelete, err := c.db.DocumentStore.GetDocumentsInTrashbin(timestamp)
	if err != nil {
		logrus.Errorf("find documents to delete: %v", err)
		return
	}
	deletedCount := 0
	for i, v := range documentsToDelete {
		logrus.Debugf("delete %d / %d documents from trashbin", i+1, len(documentsToDelete))
		err = c.deleteDocument(v)
		if err != nil {
			logrus.Errorf("cleanup trashbin: %v", err)
		} else {
			logrus.Debugf("successfully deleted document %s", v)
			deletedCount += 1
		}
	}
	if len(documentsToDelete) > 0 {
		if deletedCount > 0 {
			logrus.Infof("successfully deleted %d documents", deletedCount)
		}
		if deletedCount < len(documentsToDelete) {
			logrus.Errorf("only %d / %d documents were successfully deleted. this might need manual fixing.", deletedCount, len(documentsToDelete))
		}
	}
	logCronOp(action, deletedCount == len(documentsToDelete))
}

func (c *CronJobs) deleteDocument(docId string) error {
	err := DeleteDocument(docId)
	if err != nil {
		return fmt.Errorf("delete document %s: %v", docId, err)
	}
	err = c.db.DocumentStore.DeleteDocument(docId)
	if err != nil {
		return err
	}
	return nil
}
