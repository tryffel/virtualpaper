package process

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/services/search"
	log "tryffel.net/go/virtualpaper/util/logger"

	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

type fpConfig struct {
	id           int
	db           *storage.Database
	search       *search.Engine
	usePdfToText bool
	useOcr       bool
	usePandoc    bool
}

type fileProcessor struct {
	*Task
	taskId   string
	document *models.Document
	input    chan fileOp
	file     string
	rawFile  *os.File
	tempFile *os.File

	usePdfToText      bool
	useOcr            bool
	usePandoc         bool
	startedProcessing time.Time

	logger *logrus.Logger
	strId  string
}

func newFileProcessor(conf *fpConfig) *fileProcessor {
	fp := &fileProcessor{
		Task:  newTask(conf.id, conf.db, conf.search),
		input: make(chan fileOp, taskQueueSize),

		usePdfToText: conf.usePdfToText,
		useOcr:       conf.useOcr,
		usePandoc:    conf.usePandoc,
		strId:        fmt.Sprintf("%d", conf.id),
	}
	fp.idle = true
	fp.runFunc = fp.waitEvent
	return fp
}

func (fp *fileProcessor) logFields() logrus.Fields {
	fp.lock.RLock()
	fields := logrus.Fields{
		"module":                "process",
		"runner-id":             fp.strId,
		log.LogContextKeyTaskId: fp.taskId,
	}
	if fp.document != nil {
		fields["documentId"] = fp.document.Id
	}
	fp.lock.RUnlock()
	return fields
}

func (fp fileProcessor) Info(msg string, args ...interface{}) {
	logrus.WithFields(fp.logFields()).Infof(msg, args...)
}

func (fp fileProcessor) Debug(msg string, args ...interface{}) {
	logrus.WithFields(fp.logFields()).Debugf(msg, args...)
}

func (fp fileProcessor) Warn(msg string, args ...interface{}) {
	logrus.WithFields(fp.logFields()).Warnf(msg, args...)
}

func (fp fileProcessor) Error(msg string, args ...interface{}) {
	logrus.WithFields(fp.logFields()).Errorf(msg, args...)
}

func (fp *fileProcessor) queueFull() bool {
	return len(fp.input) == cap(fp.input)
}

func (fp *fileProcessor) queueSize() int {
	return len(fp.input)
}

func (fp *fileProcessor) GetDocumentBeingProcessed() (bool, string) {
	// this probably needs synchronization for true accuracy,
	// but it's only for metrics so it's probably okay
	fp.lock.RLock()
	defer fp.lock.RUnlock()
	doc := fp.document
	if doc == nil {
		return false, ""
	}
	return true, doc.Id
}

func (fp *fileProcessor) ProcessingDurationMs() int {
	if fp.startedProcessing.IsZero() {
		return 0
	}
	return int(time.Since(fp.startedProcessing).Milliseconds())
}

func (fp *fileProcessor) waitEvent() {
	timer := time.NewTimer(time.Millisecond * 50)
	select {
	case <-timer.C:
		// pass

	case fileOp := <-fp.input:
		defer fp.recoverPanic()
		fp.process(fileOp)
	}
}

func (fp *fileProcessor) recoverPanic() {
	// panic during processing document
	r := recover()
	if r == nil {
		return
	}

	fields := logrus.Fields{}
	fields["task_runner_id"] = fp.id
	fields[log.LogContextKeyTaskId] = fp.taskId
	if fp.document != nil {
		fields["document"] = fp.document.Id
	}
	ctx := log.ContextWithTaskId(context.Background(), fp.taskId)

	err := errors.ErrInternalError
	err.SetStack()
	err.Err = fmt.Errorf("fatal error in processing task %d: panic: %v", fp.id, r)

	fields["stack"] = string(err.Stack)
	logrus.WithFields(fields).Errorf("panic in task: %v", err.Error())

	if errors.MailEnabled() {
		msg := err.Error()
		if fp.document != nil {
			msg += fmt.Sprintf("\ndocument_id: %s\n", fp.document.Id)
		}
		err.ErrMsg = msg
		mailErr := errors.SendMail(context.Background(), err, "")
		if mailErr != nil {
			logrus.Errorf("send error stack on mail: %v", err)
		}
	}

	e := fp.cancelDocumentProcessing(ctx, "server error")
	if e != nil {
		logrus.Errorf("cancel document processing: %v", err)
	}
}

// cancel ongoing processing, in case of errors.
// without cancel processing probably gets stuck in the same processing step.
func (fp *fileProcessor) cancelDocumentProcessing(ctx context.Context, reason string) error {
	if fp.document != nil {
		log.Context(ctx).WithField("documentId", fp.document.Id).Warning("cancel processing document due to errors")
		err := fp.db.JobStore.CancelDocumentProcessing(fp.document.Id)
		if err != nil {
			logrus.Errorf("cancel document processing: %v", err)
		}
		now := time.Now()
		errDescription := fmt.Sprintf("(Processing error at %s: %s)", now.Format(time.ANSIC), reason)

		if !strings.HasPrefix(fp.document.Name, "(Error)") {
			if fp.document.Name == "" {
				fp.document.Name = "(Error) " + fp.document.Name
			} else {
				fp.document.Name = "(Error)"
			}
		}

		if !strings.HasPrefix(fp.document.Description, "(Processing error") {
			if fp.document.Description != "" {
				fp.document.Description = errDescription + "\n" + fp.document.Description
			} else {
				fp.document.Description = errDescription
			}
		}

		err = fp.db.DocumentStore.Update(storage.UserIdInternal, fp.document)
		if err != nil {
			return fmt.Errorf("update document: %v", err)
		}
		fp.document = nil
	}
	return nil
}

func (fp *fileProcessor) process(op fileOp) {
	doc, err := fp.db.DocumentStore.GetDocument(op.docId)
	if err != nil {
		logrus.Errorf("process document %s: get document: %v", op.docId, err)
		return
	}
	fp.document = doc
	fp.file = storage.DocumentPath(op.docId)

	fp.startedProcessing = time.Now()
	fp.processDocument()
	fp.startedProcessing = time.Time{}
}

// re-calculate hash. If it differs from current document.Hash, update document record and rename file to new hash,
// if different.
func (fp *fileProcessor) updateHash(ctx context.Context, doc *models.Document) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessHash,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "calculate hash")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}

	err = fp.ensureFileOpen()
	if err != nil {
		return err
	}

	defer fp.completeProcessingStep(process, job)
	hash, err := GetFileHash(fp.rawFile)
	if err != nil {
		job.Status = models.JobFailure
		return err
	}

	if hash != doc.Hash {
		log.Info(ctx, "rename file", map[string]interface{}{"old-name": doc.Hash, "new-name": hash})
	} else {
		log.Info(ctx, "hash not changed", map[string]interface{}{"name": doc.Hash})
		job.Status = models.JobFinished
		job.Message = "hash: no change"
		return nil
	}

	oldName := fp.rawFile.Name()
	err = os.Rename(oldName, path.Join(config.C.Processing.DocumentsDir, hash))
	if err != nil {
		job.Status = models.JobFailure
		return fmt.Errorf("rename file (doc %s) by old hash: %v", fp.document.Id, err)
	}

	fp.document.Hash = hash
	err = fp.db.DocumentStore.Update(storage.UserIdInternal, fp.document)
	if err != nil {
		job.Status = models.JobFailure
		return fmt.Errorf("save updated document: %v", err)
	}

	job.Status = models.JobFinished
	return nil
}

func (fp *fileProcessor) ensureFileOpen() error {
	if fp.rawFile != nil {
		return nil
	}
	var err error
	fp.rawFile, err = os.OpenFile(fp.file, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file: %v", err)
	}
	return nil
}

func (fp fileProcessor) ensureFileOpenAndLogFailure() error {
	err := fp.ensureFileOpen()
	if err != nil {
		fp.Error("open file: %v", err)
		return err
	}
	return nil
}

func (fp *fileProcessor) isDuplicate() (bool, error) {
	err := fp.ensureFileOpen()
	if err != nil {
		return false, err
	}
	hash, err := GetFileHash(fp.rawFile)
	if err != nil {
		return false, err
	}

	document, err := fp.db.DocumentStore.GetByHash(0, hash)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	if document != nil {
		return true, nil
	}
	return false, nil
}

func (fp *fileProcessor) indexSearchContent(ctx context.Context) error {
	if fp.document == nil {
		return errors.New("no document")
	}

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessFts,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "index for search engine")
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)

	if len(fp.document.Tags) == 0 {
		tags, err := fp.db.MetadataStore.GetDocumentTags(fp.document.UserId, fp.document.Id)
		if err != nil {
			if errors.Is(err, errors.ErrRecordNotFound) {
			} else {
				logrus.Errorf("get document tags: %v", err)
			}
		} else {
			fp.document.Tags = *tags
		}
	}
	if len(fp.document.Metadata) == 0 {
		metadata, err := fp.db.MetadataStore.GetDocumentMetadata(fp.document.UserId, fp.document.Id)
		if err != nil {
			if errors.Is(err, errors.ErrRecordNotFound) {
			} else {
				logrus.Errorf("get document metadata: %v", err)
			}
		} else {
			fp.document.Metadata = *metadata
		}
	}

	if fp.search == nil {
		return errors.New("no search engine available")
	}

	log.Context(ctx).Info("Send document to search index")
	err = fp.search.IndexDocuments(&[]models.Document{*fp.document}, fp.document.UserId)
	if err != nil {
		job.Message += "; " + err.Error()
		job.Status = models.JobFailure
	} else {
		job.Status = models.JobFinished
	}

	return nil
}
