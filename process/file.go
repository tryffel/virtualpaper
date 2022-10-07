package process

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/otiai10/gosseract"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/config"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/search"
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
	document *models.Document
	input    chan fileOp
	file     string
	rawFile  *os.File
	tempFile *os.File

	usePdfToText      bool
	useOcr            bool
	usePandoc         bool
	startedProcessing time.Time
}

func newFileProcessor(conf *fpConfig) *fileProcessor {
	fp := &fileProcessor{
		Task:  newTask(conf.id, conf.db, conf.search),
		input: make(chan fileOp, taskQueueSize),

		usePdfToText: conf.usePdfToText,
		useOcr:       conf.useOcr,
		usePandoc:    conf.usePandoc,
	}
	fp.idle = true
	fp.runFunc = fp.waitEvent
	return fp
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
	fields["task_id"] = fp.id
	if fp.document != nil {
		fields["document"] = fp.document.Id
	}

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
		mailErr := errors.SendMail(err, "")
		if mailErr != nil {
			logrus.Errorf("send error stack on mail: %v", err)
		}
	}

	e := fp.cancelDocumentProcessing("server error")
	if e != nil {
		logrus.Errorf("cancel document processing: %v", err)
	}
}

// cancel ongoing processing, in case of errors.
// without cancel processing probably gets stuck in the same processing step.
func (fp *fileProcessor) cancelDocumentProcessing(reason string) error {
	if fp.document != nil {
		logrus.Warningf("cancel processing document %s due to errors", fp.document.Id)
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
	if op.document == nil && op.file != "" {
		fp.file = op.file
		fp.processFile()
	} else if op.document != nil {
		fp.document = op.document
		fp.startedProcessing = time.Now()
		fp.processDocument()
		fp.startedProcessing = time.Time{}
	} else {
		logrus.Warningf("process task got empty fileop, skipping")
	}
}

func (fp *fileProcessor) processDocument() {
	logrus.Debugf("Task %d process file %s", fp.id, fp.document.Id)

	pendingSteps, err := fp.db.JobStore.GetDocumentPendingSteps(fp.document.Id)
	if err != nil {
		logrus.Errorf("get pending processing steps for document %s: %v", fp.document.Id, err)
		return
	}

	metadata, err := fp.db.MetadataStore.GetDocumentMetadata(fp.document.UserId, fp.document.Id)
	if err != nil {
		logrus.Errorf("get document metadata before processing: %v", err)
	} else {
		fp.document.Metadata = *metadata
	}

	tags, err := fp.db.MetadataStore.GetDocumentTags(fp.document.UserId, fp.document.Id)
	if err != nil {
		logrus.Errorf("get document tags before processing: %v", err)
	} else {
		fp.document.Tags = *tags
	}

	filePath := storage.DocumentPath(fp.document.Id)
	file, err := os.Open(filePath)
	if err != nil {
		logrus.Errorf("open document %s file: %v", fp.document.Id, err)
		err = fp.cancelDocumentProcessing("file not found")
		if err != nil {
			logrus.Errorf("cancel document processing: %v", err)
		}
		return
	}

	defer fp.cleanup()

	for _, step := range *pendingSteps {
		switch step.Step {
		case models.ProcessHash:
			err := fp.updateHash(fp.document, file)
			if err != nil {
				logrus.Errorf("update hash: %v", err)
				return
			} else {
				file.Close()
				file, err = os.Open(filePath)
				if err != nil {
					logrus.Errorf("open document %s file: %v", fp.document.Id, err)
					err = fp.cancelDocumentProcessing("file not found")
					if err != nil {
						logrus.Errorf("cancel document processing: %v", err)
					}
					return
				}
			}
		case models.ProcessThumbnail:
			err := fp.generateThumbnail(file)
			if err != nil {
				logrus.Errorf("generate thumbnail: %v", err)
				return
			}
		case models.ProcessParseContent:
			err := fp.parseContent(file)
			if err != nil {
				logrus.Errorf("parse content: %v", err)
				return
			}
		case models.ProcessRules:
			err := fp.runRules()
			if err != nil {
				logrus.Errorf("run rules: %v", err)
				return
			}
		case models.ProcessFts:
			err := fp.indexSearchContent()
			if err != nil {
				logrus.Errorf("index search content: %v", err)
				return
			}
		default:
			logrus.Warningf("unhandle process step: %v, skipping", step.Step)
		}
	}

	file.Close()
}

// re-calculate hash. If it differs from current document.Hash, update document record and rename file to new hash,
// if different.
func (fp *fileProcessor) updateHash(doc *models.Document, file *os.File) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessHash,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "calculate hash")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}

	defer fp.completeProcessingStep(process, job)
	hash, err := GetFileHash(file)
	if err != nil {
		job.Status = models.JobFailure
		return err
	}

	if hash != doc.Hash {
		logrus.Infof("rename file %s to %s", doc.Hash, hash)
	} else {
		logrus.Infof("file hash has not changed")
		job.Status = models.JobFinished
		job.Message = "hash: no change"
		return nil
	}

	oldName := file.Name()
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

func (fp *fileProcessor) updateThumbnail(doc *models.Document, file *os.File) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessThumbnail,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "generate thumbnail")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}
	job.Message = "Generate thumbnail"
	defer fp.completeProcessingStep(process, job)

	output := storage.PreviewPath(fp.document.Id)

	err = storage.CreatePreviewDir(fp.document.Id)
	if err != nil {
		logrus.Errorf("create preview dir: %v", err)
	}

	logrus.Infof("generate thumbnail for document %s", fp.document.Id)
	err = generateThumbnail(file.Name(), output, 0, 500, process.Document.Mimetype)

	err = fp.db.DocumentStore.Update(storage.UserIdInternal, doc)
	if err != nil {
		logrus.Errorf("update document record: %v", err)
	}

	if err != nil {
		job.Status = models.JobFailure
		job.Message += "; " + err.Error()
		return fmt.Errorf("call imagick: %v", err)
	}
	job.Status = models.JobFinished
	return nil
}

func (fp *fileProcessor) processFile() {
	logrus.Infof("task %d, process file %s", fp.id, fp.file)

	fp.lock.Lock()
	fp.idle = false
	fp.lock.Unlock()
	var err error

	fp.rawFile, err = os.OpenFile(fp.file, os.O_RDONLY, os.ModePerm)

	defer fp.cleanup()

	if err != nil {
		logrus.Errorf("process file %s, open: %v", fp.file, err)
		return
	}

	duplicate, err := fp.isDuplicate()
	if duplicate {
		logrus.Infof("file %s is a duplicate, ignore file", fp.file)
		return
	}

	if err != nil {
		logrus.Errorf("get duplicate status: %v", err)
		return
	}

	err = fp.createNewDocumentRecord()
	if err != nil {
		logrus.Error(err)
		return
	}

	logrus.Info("generate thumbnail")
	err = fp.generateThumbnail(fp.rawFile)
	if err != nil {
		logrus.Errorf("generate thumbnail: %v", err)
		return
	}

	logrus.Info("parse content")
	err = fp.parseContent(fp.rawFile)
	if err != nil {
		logrus.Errorf("Parse document content: %v", err)
	}
}

func (fp *fileProcessor) cleanup() {
	logrus.Infof("Stop processing file %s", fp.file)

	if fp.rawFile != nil {
		fp.rawFile.Close()
		fp.rawFile = nil
	}
	if fp.tempFile != nil {
		fp.tempFile.Close()

		err := os.Remove(fp.tempFile.Name())
		if err != nil {
			logrus.Errorf("remove temp file %s: %v", fp.tempFile.Name(), err)
		}
		fp.tempFile = nil
	}

	if fp.document != nil {
		tmpDir := storage.TempFilePath(fp.document.Hash)
		err := os.RemoveAll(tmpDir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
			} else {
				logrus.Errorf("cannot remove tmp dir %s: %v", tmpDir, err)
			}
		}
	}

	fp.document = nil
	fp.file = ""
	fp.lock.Lock()
	fp.idle = true
	fp.lock.Unlock()
}

func (fp *fileProcessor) isDuplicate() (bool, error) {
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

func (fp *fileProcessor) createNewDocumentRecord() error {
	fullDir, fileName := path.Split(fp.file)
	fullDir = strings.TrimSuffix(fullDir, "/")
	fullDir = strings.TrimSuffix(fullDir, "\\")
	_, userName := path.Split(fullDir)
	userName = strings.Trim(userName, "/\\")

	user, err := fp.db.UserStore.GetUserByName(userName)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return fmt.Errorf("unable to process document from input dir, since user '%s' does not exist. Ensure "+
				"user has properly named directory assigned to them, and add documents there", userName)
		} else {
			return fmt.Errorf("get user: %v", err)
		}
	}

	doc := &models.Document{
		UserId:   user.Id,
		Name:     fileName,
		Content:  "",
		Filename: fileName,
		Date:     time.Now(),
	}
	doc.UpdatedAt = time.Now()
	doc.CreatedAt = time.Now()

	doc.Hash, err = GetFileHash(fp.rawFile)
	if err != nil {
		return fmt.Errorf("get hash: %v", err)
	}

	err = fp.db.DocumentStore.Create(doc)
	if err != nil {
		return fmt.Errorf("store document: %s", err)
	}

	fp.document = doc
	return nil
}

func (fp *fileProcessor) generateThumbnail(file *os.File) error {
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessThumbnail,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "generate thumbnail")
	if err != nil {
		return fmt.Errorf("persist process item: %v", err)
	}
	defer fp.completeProcessingStep(process, job)

	output := storage.PreviewPath(fp.document.Id)
	err = storage.CreatePreviewDir(fp.document.Id)
	if err != nil {
		return fmt.Errorf("create thumbnail output dir: %v", err)
	}

	name := file.Name()
	err = generateThumbnail(name, output, 0, 500, fp.document.Mimetype)
	if err != nil {
		job.Status = models.JobFailure
		job.Message += "; " + err.Error()
		return fmt.Errorf("call imagick: %v", err)
	}

	job.Status = models.JobFinished
	return nil
}

func (fp *fileProcessor) parseContent(file *os.File) error {

	logrus.Infof("extract content for document %s", fp.document.Id)
	if fp.document.IsPdf() {
		return fp.extractPdf(file)
	} else if fp.document.IsImage() {
		return fp.extractImage(file)
	} else if fp.usePandoc && isPandocMimetype(fp.document.Mimetype) {
		return fp.extractPandoc(file)
	} else {
		return fmt.Errorf("cannot extract content from mimetype: %v", fp.document.Mimetype)
	}
}

func (fp *fileProcessor) extractPdf(file *os.File) error {
	// if pdf, generate image preview and pass it to tesseract
	var err error

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Document:   nil,
		Step:       models.ProcessParseContent,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "extract pdf content")
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)

	var text string
	useOcr := false

	if fp.usePdfToText {
		logrus.Infof("Attempt to parse document %s content with pdftotext", fp.document.Id)
		text, err = getPdfToText(file, fp.document.Id)
		if err != nil {
			if err.Error() == "empty" {
				logrus.Infof("document %s has no plain text, try ocr", fp.document.Id)
				useOcr = true
			} else {
				logrus.Debugf("failed to get content with pdftotext: %v", err)
			}
		} else {
			useOcr = false
		}
	} else {
		useOcr = true
	}

	if useOcr {
		text, err = runOcr(file.Name(), fp.document.Id)
		if err != nil {
			job.Message += "; " + err.Error()
			job.Status = models.JobFailure
			return fmt.Errorf("parse document content: %v", err)
		}
	}

	if text == "" {
		logrus.Warningf("document %s content seems to be empty", fp.document.Id)
	}

	text = strings.ToValidUTF8(text, "")

	fp.document.Content = text
	err = fp.db.DocumentStore.SetDocumentContent(fp.document.Id, text)
	if err != nil {
		job.Message += "; " + "save document content: " + err.Error()
		job.Status = models.JobFailure
		return fmt.Errorf("save document content: %v", err)
	} else {
		job.Status = models.JobFinished
	}
	return nil
}

func (fp *fileProcessor) extractImage(file *os.File) error {
	var err error
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessParseContent,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "extract content from image")
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)
	client := gosseract.NewClient()
	defer client.Close()

	err = client.SetImage(file.Name())
	if err != nil {
		return fmt.Errorf("set ocr image source: %v", err)
	}
	text, err := client.Text()
	if err != nil {
		job.Message += "; " + err.Error()
		job.Status = models.JobFailure
		return fmt.Errorf("parse document text: %v", err)
	} else {
		text = strings.ToValidUTF8(text, "")
		fp.document.Content = text
		err = fp.db.DocumentStore.SetDocumentContent(fp.document.Id, fp.document.Content)
		if err != nil {
			job.Message += "; " + "save document content: " + err.Error()
			job.Status = models.JobFailure
			return fmt.Errorf("save document content: %v", err)
		} else {
			job.Status = models.JobFinished
		}
	}
	return nil
}

func (fp *fileProcessor) extractPandoc(file *os.File) error {
	var err error
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessParseContent,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, fmt.Sprintf("extract content from %s", fp.document.Mimetype))
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)

	text, err := getPandocText(fp.document.Mimetype, fp.document.Filename, file)
	if err != nil {
		job.Message += "; " + err.Error()
		job.Status = models.JobFailure
		return fmt.Errorf("parse document text: %v", err)
	} else {
		text = strings.ToValidUTF8(text, "")
		fp.document.Content = text
		err = fp.db.DocumentStore.SetDocumentContent(fp.document.Id, fp.document.Content)
		if err != nil {
			job.Message += "; " + "save document content: " + err.Error()
			job.Status = models.JobFailure
			return fmt.Errorf("save document content: %v", err)
		} else {
			job.Status = models.JobFinished
		}
	}
	return nil
}

func (fp *fileProcessor) runRules() error {
	if fp.document == nil {
		return errors.New("no document set")
	}

	rules, err := fp.db.RuleStore.GetActiveUserRules(fp.document.UserId)
	if err != nil {
		return fmt.Errorf("load rules: %v", err)
	}

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessRules,
		CreatedAt:  time.Now(),
	}
	job, err := fp.db.JobStore.StartProcessItem(process, "process user rules")
	// hotfix for failure when job item does not exist anymore.
	if err != nil {
		logrus.Warningf("persist job record: %v", err)
		// use empty job to not panic the rest of the function
		job = &models.Job{}
	} else {
		defer fp.completeProcessingStep(process, job)
	}

	metadataValues, err := fp.db.MetadataStore.GetUserValuesWithMatching(fp.document.UserId)
	if err != nil {
		logrus.Errorf("get metadata values with matching for user %d: %v", fp.document.UserId, err)
	} else if len(*metadataValues) != 0 {
		err = matchMetadata(fp.document, metadataValues)
	}

	for i, rule := range rules {
		logrus.Debugf("(%d.) run user rule %d", i, rule.Id)

		if len(rule.Actions) == 0 {
			logrus.Debugf("rule %d does not have actions, skip rule", rule.Id)
			continue
		}

		if len(rule.Conditions) == 0 {
			logrus.Debugf("rule %d does not have conditions, skip rule", rule.Id)
			continue
		}

		runner := NewDocumentRule(fp.document, rule)
		match, err := runner.Match()
		if err != nil {
			logrus.Errorf("match rule (%d): %v", rule.Id, err)
		}
		if !match {
			logrus.Debugf("document %s does not match rule: %d", fp.document.Id, rule.Id)
		} else {

			logrus.Debugf("document %s matches rule %d, run actions", fp.document.Id, rule.Id)
			err = runner.RunActions()
			if err != nil {
				logrus.Errorf("rule (%d) actions: %v", rule.Id, err)
			}
		}
	}

	if err != nil {
		logrus.Errorf("run user rules: %v", err)
		job.Status = models.JobFailure
	} else {
		job.Status = models.JobFinished
	}

	err = fp.db.DocumentStore.Update(storage.UserIdInternal, fp.document)
	if err != nil {
		logrus.Errorf("update document (%s) after rules: %v", fp.document.Id, err)
	}

	metadata := make([]models.Metadata, len(fp.document.Metadata))
	for i, _ := range fp.document.Metadata {
		metadata[i] = fp.document.Metadata[i]
	}
	err = fp.db.MetadataStore.UpdateDocumentKeyValues(fp.document.UserId, fp.document.Id, metadata)
	if err != nil {
		logrus.Errorf("update document metadata after processing rules")
	} else {
		// metadata added by rule does not contain all fields, only key/value ids. Load other values as well.
		newMetadata, err := fp.db.MetadataStore.GetDocumentMetadata(fp.document.UserId, fp.document.Id)
		if err != nil {
			logrus.Errorf("reload full metadata records for document "+
				"after (doc %s) rules: %v", fp.document.Id, err)
		} else {
			fp.document.Metadata = *newMetadata
		}
	}
	return nil
}

func (fp *fileProcessor) indexSearchContent() error {
	if fp.document == nil {
		return errors.New("no document")
	}

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Step:       models.ProcessFts,
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

	err = fp.search.IndexDocuments(&[]models.Document{*fp.document}, fp.document.UserId)
	if err != nil {
		job.Message += "; " + err.Error()
		job.Status = models.JobFailure
	} else {
		job.Status = models.JobFinished
	}

	return nil
}

func (fp *fileProcessor) completeProcessingStep(process *models.ProcessItem, job *models.Job) {

	// remove step if it was successful. In addition, remove step from queue.
	// if further steps do not absolutely require running this step.
	removeStep := job.Status == models.JobFinished
	switch process.Step {
	case models.ProcessThumbnail:
		removeStep = true
	case models.ProcessRules:
		removeStep = true
	case models.ProcessFts:
		removeStep = true
	}

	if job.Status == models.JobFailure && removeStep {
		logrus.Infof("failure in processing document %s, skipping step %s", job.DocumentId, job.Step.String())
		// prevent 100% cpu utilization if step fails
		time.Sleep(1)
	}

	err := fp.db.JobStore.MarkProcessingDone(process, removeStep)
	if err != nil {
		logrus.Errorf("mark process complete: %v", err)
	}
	job.StoppedAt = time.Now()
	if job.Status == models.JobRunning {
		job.Status = models.JobFailure
	}
	err = fp.db.JobStore.Update(job)
	if err != nil {
		logrus.Errorf("save job to database: %v", err)
	}
}

//DeleteDocument deletes original document and its preview file.
func DeleteDocument(docId string) error {
	previewPath := storage.PreviewPath(docId)
	docPath := storage.DocumentPath(docId)

	logrus.Debugf("delete preview file %s", previewPath)
	err := os.Remove(previewPath)
	if err != nil {
		return fmt.Errorf("remove thumbnail: %v", err)
	}
	logrus.Debugf("delete document file %s", previewPath)
	err = os.Remove(docPath)
	if err != nil {
		return fmt.Errorf("remove document file: %v", err)
	}
	return nil
}
