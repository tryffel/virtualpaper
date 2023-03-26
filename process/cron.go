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
