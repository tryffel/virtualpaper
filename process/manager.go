package process

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	config "tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

const idleCheckDocumentsForProcessingSec = time.Second * 5
const taskQueueSize = 100

// processing file operation. Either fill file or document.
// Fill file to mark file without document record.
// File document to mark existing document.
type fileOp struct {
	file     string
	document *models.Document
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

	inputWatch *fsnotify.Watcher

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
	manager.inputWatch, err = fsnotify.NewWatcher()
	return manager, err
}

func (m *Manager) Start() error {
	if m.isRunning() {
		return errors.New("already running")
	}

	logrus.Info("Start background worker")
	logrus.Infof("Watch directory %s", config.C.Processing.InputDir)

	if config.C.Processing.InputDir != "" {
		logrus.Infof("add directory wath for %s", config.C.Processing.InputDir)

		err := filepath.Walk(config.C.Processing.InputDir, func(filepath string, info os.FileInfo, err error) error {
			logrus.Debugf("add dir watch for: %s", filepath)
			err = m.inputWatch.Add(filepath)
			if err != nil {
				_, file := path.Split(filepath)
				return fmt.Errorf(file)
			}
			return err
		})
		if err != nil {
			logrus.Errorf("add input watch: %v", err)
		}
	}

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
		m.inputWatch.Close()
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

	docs := make(map[string]*models.Document)
	processes, _, err := m.db.JobStore.GetPendingProcessing()

	if len(*processes) > 0 {
		logrus.Infof("%d more document(s) are waiting for processing, push steps to runners", len(*processes))
	}

	if err != nil {
		logrus.Errorf("get pending processing: %v", err)
	} else {
		for _, v := range *processes {
			if docs[v.DocumentId] == nil {
				doc, err := m.db.DocumentStore.GetDocument(0, v.DocumentId)
				if err != nil {
					logrus.Errorf("get document: %v", err)
				} else {

					metadata, err := m.db.MetadataStore.GetDocumentMetadata(0, v.DocumentId)
					if err != nil {
						logrus.Errorf("get documetn metadata: %v", err)
					} else {
						doc.Metadata = *metadata
						docs[v.DocumentId] = doc
					}
				}
			}
		}

		for _, v := range docs {
			err = m.AddDocumentForProcessing(v)
			if err != nil {
				logrus.Errorf("add document for processing: %v", err)
			}
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
func (m *Manager) AddDocumentForProcessing(doc *models.Document) error {
	filePath := storage.DocumentPath(doc.Id)
	if !m.QueueFull() {
		m.scheduleNewOp(filePath, doc)
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

	case event, ok := <-m.inputWatch.Events:
		if ok {
			logrus.Infof("Got file watcher event: %v", event)
		}

		if event.Op == fsnotify.Write {
			logrus.Infof("Schedule processing for file %s", event.Name)
			m.scheduleNewOp(event.Name, nil)
		}
		m.PullDocumentsToProcess()
	case report := <-m.reportChan:
		logrus.Infof("Got task report: %v", report)

	}
	time.Sleep(time.Second)
}

// schedule file operation to any idle task. If none of the tasks are idle, queue it to random task.
func (m *Manager) scheduleNewOp(file string, doc *models.Document) {
	if doc != nil {
		logrus.Debugf("schedule new file process for document %s", doc.Id)
	} else {
		logrus.Debugf("schedule new file process for file %s", file)

	}
	op := fileOp{file: file, document: doc}
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
