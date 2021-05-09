package process

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"gopkg.in/gographics/imagick.v3/imagick"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
	config "tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/search"
	"tryffel.net/go/virtualpaper/storage"
)

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
}

func NewManager(database *storage.Database, search *search.Engine) (*Manager, error) {
	manager := &Manager{
		lock:       &sync.RWMutex{},
		reportChan: make(chan TaskReport, 10),
		db:         database,
		search:     search,
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
		time.Sleep(time.Millisecond * 3000)

		docs := map[string]*models.Document{}
		processes, _, err := m.db.JobStore.GetPendingProcessing()
		if err != nil {
			logrus.Errorf("get pending processing: %v", err)
		} else {
			for _, v := range *processes {
				if docs[v.DocumentId] == nil {
					doc, err := m.db.DocumentStore.GetDocument(0, v.DocumentId)
					if err != nil {
						logrus.Errorf("get document: %v", err)
					} else {
						docs[v.DocumentId] = doc
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

	for _, task := range m.tasks {
		task.Stop()
	}

	m.lock.Lock()
	m.running = false
	m.lock.Unlock()
	return nil
}

// AddDocumentForProcessing marks document as available for processing.
func (m *Manager) AddDocumentForProcessing(doc *models.Document) error {
	filePath := storage.DocumentPath(doc.Id)
	m.scheduleNewOp(filePath, doc)
	return nil
}

func (m *Manager) isRunning() bool {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.running
}

// async function loop to wait for events and launch tasks.
func (m *Manager) runFunc() {
	timer := time.NewTimer(time.Millisecond * 100)

	select {
	case <-timer.C:
		// pass

	case event, ok := <-m.inputWatch.Events:
		if ok {
			logrus.Infof("Got file watcher event: %v", event)
		}

		if event.Op == fsnotify.Write {
			logrus.Infof("Schedule processing for file %s", event.Name)
			m.scheduleNewOp(event.Name, nil)
		}

		//pass

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
	scheduled := false

	for _, task := range m.tasks {
		if task.isIdle() {
			task.input <- op
			scheduled = true
			break
		}
	}

	if !scheduled {
		id := rand.Intn(m.numtasks)
		m.tasks[id].input <- op
	}
}

func Init() {
	logrus.Debugf("Initialize imagick instance")
	imagick.Initialize()
}

func Deinit() {
	logrus.Debugf("Release imagick instance")
	imagick.Terminate()
}
