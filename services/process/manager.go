package process

import (
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

const idleCheckDocumentsForProcessingSec = time.Second * 5
const taskQueueSize = 100

// processing file operation. Either fill file or document.
// Fill file to mark file without document record.
// File document to mark existing document.
type fileOp struct {
	docId string
}

// Manager manages multiple goroutines processing files.
type Manager struct {
	lock       *sync.RWMutex
	running    bool
	reportChan chan TaskReport
	db         *storage.Database
	search     *search.Engine

	tasks    []*fileProcessor
	numtasks int

	checkJobstimer *time.Timer
	runFunctimer   *time.Timer
}

func NewManager(database *storage.Database, search *search.Engine) (*Manager, error) {
	manager := &Manager{
		lock:           &sync.RWMutex{},
		reportChan:     make(chan TaskReport, 10),
		db:             database,
		search:         search,
		checkJobstimer: time.NewTimer(idleCheckDocumentsForProcessingSec),
		runFunctimer:   time.NewTimer(time.Millisecond * 100),
	}

	useOcr := true
	usePdfToText := true

	err := testPdfToText()
	if err != nil {
		usePdfToText = false
	}

	usePandoc := true
	err = testPandoc()
	if err != nil {
		usePandoc = false
	}

	buildMimeDataMapping()

	count := config.C.Processing.MaxWorkers
	manager.numtasks = count
	manager.tasks = make([]*fileProcessor, count)

	for i := 0; i < count; i++ {
		conf := &fpConfig{
			id:           i,
			db:           database,
			search:       search,
			usePdfToText: usePdfToText,
			useOcr:       useOcr,
			usePandoc:    usePandoc,
		}
		manager.tasks[i] = newFileProcessor(conf)
	}
	return manager, err
}

func (m *Manager) Start() error {
	if config.C.Processing.Disabled {
		logrus.Warningf("processing disabled, refuse to start process manager")
		return nil
	}

	if m.isRunning() {
		return errors.New("already running")
	}

	logrus.Info("Start background worker")
	for _, task := range m.tasks {
		task.Start()
	}

	f := func() {
		m.lock.Lock()
		m.running = true
		m.lock.Unlock()
		logrus.Debug("start background task manager")

		err := m.db.JobStore.CancelRunningProcesses()
		if err != nil {
			logrus.Errorf("cancel old processes: %v", err)
		}
		time.Sleep(time.Millisecond * 5000)
		m.PullDocumentsToProcess()

		for m.isRunning() {
			m.runFunc()
		}
		logrus.Debug("background task manager stopped")
	}

	go f()
	return nil
}

func (m *Manager) Stop() error {
	if !m.isRunning() {
		return errors.New("not running")
	}

	logrus.Info("Stop process manager")

	for _, task := range m.tasks {
		task.Stop()
	}

	m.lock.Lock()
	m.running = false
	m.lock.Unlock()
	return nil
}

func (m *Manager) PullDocumentsToProcess() {
	if m.QueueFull() {
		logrus.Warn("processing queue is full, don't pull more jobs yet")
		return
	}
	docs, err := m.db.JobStore.GetDocumentsPendingProcessing()
	if err != nil {
		logrus.Errorf("get documents pending for processing: %v", err)
		return
	}
	if len(*docs) == 0 {
		logrus.Debugf("no documents to process")
	}

	if len(*docs) > 0 {
		logrus.Infof("push %d documents for processing runners", len(*docs))
	}

	for _, v := range *docs {
		err = m.AddDocumentForProcessing(v)
		if err != nil {
			logrus.Errorf("add document for processing: %v", err)
		}
	}
}

func (m *Manager) QueueFull() bool {
	for _, v := range m.tasks {
		if !v.queueFull() {
			return false
		}
	}
	return true
}

type QueueStatus struct {
	TaskId               int    `json:"task_id"`
	QueueCapacity        int    `json:"queue_capacity"`
	Queued               int    `json:"queued"`
	ProcessingOngoing    bool   `json:"processing_ongoing"`
	ProcessingDocumentId string `json:"processing_document_id"`
	Running              bool   `json:"task_running"`
	DurationMs           int    `json:"duration_ms"`
}

func (m *Manager) ProcessingStatus() []QueueStatus {
	status := make([]QueueStatus, m.numtasks)
	for i, v := range m.tasks {
		status[i].TaskId = v.id
		status[i].QueueCapacity = taskQueueSize
		status[i].Queued = v.queueSize()
		status[i].ProcessingOngoing, status[i].ProcessingDocumentId = v.GetDocumentBeingProcessed()
		status[i].Running = v.isRunning()
		status[i].DurationMs = v.ProcessingDurationMs()
	}
	return status
}

// AddDocumentForProcessing marks document as available for processing.
func (m *Manager) AddDocumentForProcessing(docId string) error {
	if !m.QueueFull() {
		m.scheduleNewOp(docId)
	}
	return nil
}

func (m *Manager) isRunning() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.running
}

// async function loop to wait for events and launch tasks.
func (m *Manager) runFunc() {

	select {
	case <-m.runFunctimer.C:
		m.runFunctimer.Reset(time.Millisecond * 100)

	case <-m.checkJobstimer.C:
		m.PullDocumentsToProcess()
		m.checkJobstimer.Reset(idleCheckDocumentsForProcessingSec)

	case report := <-m.reportChan:
		logrus.Infof("Got task report: %v", report)

	}
	time.Sleep(time.Second)
}

// schedule file operation to any idle task. If none of the tasks are idle, queue it to random task.
func (m *Manager) scheduleNewOp(docId string) {
	op := fileOp{docId: docId}
	leastQueuedTask := m.tasks[0]
	for _, v := range m.tasks {
		if ok, _ := v.GetDocumentBeingProcessed(); !ok {
			// idle, pick
			leastQueuedTask = v
			break
		}

		if v.queueSize() < leastQueuedTask.queueSize() {
			leastQueuedTask = v
		}
	}
	leastQueuedTask.input <- op
}
