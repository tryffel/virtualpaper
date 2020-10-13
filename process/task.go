package process

import (
	"errors"
	"github.com/sirupsen/logrus"
	"sync"
	"tryffel.net/go/virtualpaper/storage"
)

type taskStatus int

const (
	statusError taskStatus = iota
	statusUpdate
	statusFinished
)

type TaskReport struct {
	taskId int
	status taskStatus
}

// Task is a background worker. Define runFunc and run it with Start(). Task calls runFunc in a loop
// until Stop() is called. Task can communicate via report channel, which is up to user to set correctly.
type Task struct {
	lock    *sync.RWMutex
	running bool
	idle    bool
	id      int
	db      *storage.Database
	report  *chan TaskReport

	runFunc func()
}

func newTask(id int, db *storage.Database) *Task {
	task := &Task{
		id:   id,
		lock: &sync.RWMutex{},
		db:   db,
	}
	return task
}

func (t *Task) Start() error {
	if t.isRunning() {
		return errors.New("already running")
	}

	if t.runFunc == nil {
		return errors.New("no running function defined")
	}

	f := func() {
		t.lock.Lock()
		t.running = true
		t.lock.Unlock()
		logrus.Debugf("start background task %d", t.id)

		for t.isRunning() {
			t.runFunc()
		}
		logrus.Debugf("background task %d stopped", t.id)
	}
	go f()
	return nil
}

func (t *Task) Stop() error {
	if !t.isRunning() {
		return errors.New("not running")
	}

	t.lock.Lock()
	t.running = false
	t.lock.Unlock()
	return nil
}

func (t *Task) isRunning() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.running
}

func (t *Task) isIdle() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.idle
}

func (t *Task) emitReport(status taskStatus) error {
	if t.report == nil {
		return errors.New("no report channel")
	}

	*t.report <- TaskReport{
		taskId: t.id,
		status: status,
	}

	return nil
}
