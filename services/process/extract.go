package process

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
	"tryffel.net/go/virtualpaper/models"
)

func (fp *fileProcessor) parseContent(ctx context.Context) error {
	err := fp.ensureFileOpen()
	if err != nil {
		logrus.Errorf("open file: %v", err)

	}
	file := fp.rawFile

	fp.Info("extract content for document %s", fp.document.Id)
	if fp.document.IsPdf() {
		return fp.extractPdf(ctx, file)
	} else if fp.document.IsImage() {
		return fp.extractImage(ctx, file)
	} else if fp.usePandoc && isPandocMimetype(fp.document.Mimetype) {
		return fp.extractPandoc(ctx, file)
	} else {
		return fmt.Errorf("cannot extract content from mimetype: %v", fp.document.Mimetype)
	}
}

func (fp *fileProcessor) extractPdf(ctx context.Context, file *os.File) error {
	// if pdf, generate image preview and pass it to tesseract
	var err error

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Document:   nil,
		Action:     models.ProcessParseContent,
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
		fp.Info("Attempt to parse document content with pdftotext")
		text, err = getPdfToText(file, fp.document.Id)
		if err != nil {
			if err.Error() == "empty" {
				fp.Info("document has no plain text, try ocr")
				useOcr = true
			} else {
				fp.Debug("failed to get content with pdftotext: %v", err)
			}
		} else {
			useOcr = false
		}
	} else {
		useOcr = true
	}

	if useOcr {
		text, err = runOcr(ctx, file.Name(), fp.document.Id)
		if err != nil {
			job.Message += "; " + err.Error()
			job.Status = models.JobFailure
			return fmt.Errorf("parse document content: %v", err)
		}
	}

	if text == "" {
		fp.Warn("content seems to be empty")
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

func (fp *fileProcessor) extractImage(ctx context.Context, file *os.File) error {
	var err error
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessParseContent,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, "extract content from image")
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)

	text, err := runOcr(ctx, file.Name(), fp.document.Id)
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

func (fp *fileProcessor) extractPandoc(ctx context.Context, file *os.File) error {
	var err error
	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessParseContent,
		CreatedAt:  time.Now(),
	}

	job, err := fp.db.JobStore.StartProcessItem(process, fmt.Sprintf("extract content from %s", fp.document.Mimetype))
	if err != nil {
		return fmt.Errorf("start process: %v", err)
	}

	defer fp.completeProcessingStep(process, job)

	text, err := getPandocText(ctx, fp.document.Mimetype, fp.document.Filename, file)
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
