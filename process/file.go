package process

import (
	"errors"
	"fmt"
	"github.com/otiai10/gosseract"
	"github.com/sirupsen/logrus"
	"gopkg.in/gographics/imagick.v3/imagick"
	"os"
	"path"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/config"
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
}

type fileProcessor struct {
	*Task
	document *models.Document
	input    chan fileOp
	file     string
	rawFile  *os.File
	tempFile *os.File

	usePdfToText bool
	useOcr       bool
}

func newFileProcessor(conf *fpConfig) *fileProcessor {
	fp := &fileProcessor{
		Task:  newTask(conf.id, conf.db, conf.search),
		input: make(chan fileOp, 5),

		usePdfToText: conf.usePdfToText,
		useOcr:       conf.useOcr,
	}
	fp.idle = true
	fp.runFunc = fp.waitEvent
	return fp
}

func (fp *fileProcessor) waitEvent() {
	timer := time.NewTimer(time.Millisecond * 50)
	select {
	case <-timer.C:
		// pass

	case fileOp := <-fp.input:
		fp.process(fileOp)
		//fp.processFile()

		//fp.processFile()
	}
}

func (fp *fileProcessor) process(op fileOp) {
	if op.document == nil && op.file != "" {
		fp.processFile()
	} else if op.document != nil {
		fp.document = op.document
		fp.processDocument()
	} else {
		logrus.Warningf("process task got empty fileop, skipping")
	}
}

func (fp *fileProcessor) processDocument() {

	pendingSteps, err := fp.db.JobStore.GetDocumentPendingSteps(fp.document.Id)
	if err != nil {
		logrus.Errorf("get pending processing steps for document %d: %v", fp.document.Id, err)
		return
	}

	logrus.Debugf("Task %d process file %d", fp.id, fp.document.Id)

	file, err := os.Open(path.Join(config.C.Processing.DocumentsDir, fp.document.Hash))
	if err != nil {
		logrus.Errorf("open document %d file: %v", fp.document.Id, err)
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
				file, err = os.Open(path.Join(config.C.Processing.DocumentsDir, fp.document.Hash))
				if err != nil {
					logrus.Errorf("open document %d file: %v", fp.document.Id, err)
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

	defer fp.persistProcess(process, job)
	hash, err := getHash(file)
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
		return fmt.Errorf("rename file (doc %d) by old hash: %v", fp.document.Id, err)
	}

	fp.document.Hash = hash
	err = fp.db.DocumentStore.Update(fp.document)
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
	defer fp.persistProcess(process, job)

	output := path.Join(config.C.Processing.PreviewsDir, fp.document.Hash+".png")

	logrus.Infof("generate thumbnail for document %d", fp.document.Id)
	_, err = imagick.ConvertImageCommand([]string{
		"convert", "-thumbnail", "x500", "-background", "white", "-alpha", "remove", file.Name() + "[0]", output,
	})

	err = fp.db.DocumentStore.Update(doc)
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
		tmpDir := path.Join(config.C.Processing.TmpDir, fp.document.Hash)
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
	hash, err := getHash(fp.rawFile)
	if err != nil {
		return false, err
	}

	document, err := fp.db.DocumentStore.GetByHash(hash)
	if err != nil {
		if errors.Is(err, storage.ErrRecordNotFound) {
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
	_, fileName := path.Split(fp.file)

	doc := &models.Document{
		UserId:   5,
		Name:     fileName,
		Content:  "",
		Filename: fileName,
		Date:     time.Now(),
	}
	doc.UpdatedAt = time.Now()
	doc.CreatedAt = time.Now()

	var err error
	doc.Hash, err = getHash(fp.rawFile)
	if err != nil {
		return fmt.Errorf("get hash: %v", err)
	}

	err = fp.db.DocumentStore.Create(doc)
	if err != nil {
		return fmt.Errorf("store document: %v", err)
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
	defer fp.persistProcess(process, job)

	output := path.Join(config.C.Processing.PreviewsDir, fp.document.Hash+".png")

	name := file.Name()
	err = generateThumbnail(name, output, 0, 500)
	if err != nil {
		job.Status = models.JobFailure
		job.Message += "; " + err.Error()
		return fmt.Errorf("call imagick: %v", err)
	}

	job.Status = models.JobFinished
	return nil
}

func (fp *fileProcessor) parseContent(file *os.File) error {

	logrus.Infof("extract content for document %d", fp.document.Id)
	if fp.document.IsPdf() {
		return fp.extractPdf(file)
	} else if fp.document.IsImage() {
		return fp.extractImage(file)
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

	defer fp.persistProcess(process, job)

	var text string
	useOcr := false

	if fp.usePdfToText {
		logrus.Infof("Attempt to parse document %d content with pdftotext", fp.document.Id)
		text, err = getPdfToText(file, fp.document.Hash)
		if err != nil {
			if err.Error() == "empty" {
				logrus.Infof("document %d has no plain text, try ocr", fp.document.Id)
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
		text, err = runOcr(file.Name(), fp.document.Hash)
		if err != nil {
			job.Message += "; " + err.Error()
			job.Status = models.JobFailure
			return fmt.Errorf("parse document content: %v", err)
		}
	}

	if text == "" {
		logrus.Warningf("document %d content seems to be empty", fp.document.Id)
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

	defer fp.persistProcess(process, job)
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

	defer fp.persistProcess(process, job)

	if len(fp.document.Tags) == 0 {
		tags, err := fp.db.MetadataStore.GetDocumentTags(fp.document.UserId, fp.document.Id)
		if err != nil {
			if errors.Is(err, storage.ErrRecordNotFound) {
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
			if errors.Is(err, storage.ErrRecordNotFound) {
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

func (fp *fileProcessor) persistProcess(process *models.ProcessItem, job *models.Job) {
	err := fp.db.JobStore.MarkProcessingDone(process, job.Status == models.JobFinished)
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

func (fp *fileProcessor) persistJob(job *models.Job) {
	job.StoppedAt = time.Now()
	err := fp.db.JobStore.Update(job)
	if err != nil {
		logrus.Errorf("save job to database: %v", err)
	}
}
