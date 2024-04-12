package process

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
	log "tryffel.net/go/virtualpaper/util/logger"
)

func (fp *fileProcessor) completeProcessingStep(process *models.ProcessItem, job *models.Job) {
	fp.Debug("processing completed, status: %v", job.Status)

	// remove step if it was successful. In addition, remove step from queue.
	// if further steps do not absolutely require running this step.
	removeStep := job.Status == models.JobFinished
	switch process.Action {
	case models.ProcessThumbnail, models.ProcessDetectLanguage, models.ProcessRules, models.ProcessFts:
		removeStep = true
	}

	if job.Status == models.JobFailure && removeStep {
		fp.Info("failure in processing document %s, skipping step %s", job.DocumentId, job.Step.String())
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
	err = fp.db.JobStore.UpdateJob(job)
	if err != nil {
		logrus.Errorf("save job to database: %v", err)
	}
}

func (fp *fileProcessor) cleanup() {
	fp.Info("stop processing file")

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

/* New implementation, used when document is sent using the API */
func (fp *fileProcessor) processDocument() {
	fp.Info("Start processing file")

	refreshDocument := func() error {
		doc, err := fp.db.DocumentStore.GetDocument(fp.document.Id)
		if err != nil {
			return fmt.Errorf("get document: %v", err)
		}
		metadata, err := fp.db.MetadataStore.GetDocumentMetadata(0, fp.document.Id)
		if err != nil {
			return fmt.Errorf("get metadata: %v", err)
		}
		doc.Metadata = *metadata
		fp.document = doc
		return nil
	}

	defer fp.cleanup()
	for {
		fp.taskId, _ = uuid.GenerateUUID()
		ctx := log.ContextWithTaskId(context.Background(), fp.taskId)
		step, err := fp.db.JobStore.GetNextStepForDocument(fp.document.Id)
		if err != nil {
			if errors.Is(err, errors.ErrRecordNotFound) {
				// all steps executed
			} else {
				logrus.Errorf("get next processing step for document %s: %v", fp.document.Id, err)
			}
			break
		}
		fp.Info("run step %s", step.Action)

		switch step.Action {
		case models.ProcessHash:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing(ctx, "file not found")
				if err != nil {
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.updateHash(ctx, fp.document)
			if err != nil {
				log.Errorf(ctx, "update hash %v", err)
				return
			}
		case models.ProcessThumbnail:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing(ctx, "file not found")
				if err != nil {
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.generateThumbnail(ctx)
			if err != nil {
				logrus.Errorf("generate thumbnail: %v", err)
				return
			}
		case models.ProcessParseContent:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing(ctx, "file not found")
				if err != nil {
					log.Errorf(ctx, "cancel document : %v", err)
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.parseContent(ctx)
			if err != nil {
				log.Errorf(ctx, "parse content: %v", err)
				return
			}
		case models.ProcessDetectLanguage:
			err := refreshDocument()
			if err != nil {
				log.Errorf(ctx, "refresh document: %v", err)
				return
			}
			err = fp.detectLanguage(ctx)
			if err != nil {
				log.Errorf(ctx, "detect language: %v", err)
				return
			}
		case models.ProcessRules:
			err := refreshDocument()
			if err != nil {
				log.Errorf(ctx, "refresh document: %v", err)
				return
			}
			err = fp.runRules(ctx, step.Trigger)
			if err != nil {
				log.Errorf(ctx, "run rules: %v", err)
				return
			}
		case models.ProcessFts:
			err := refreshDocument()
			if err != nil {
				log.Errorf(ctx, "refresh document: %v", err)
				return
			}
			err = fp.indexSearchContent(ctx)
			if err != nil {
				log.Errorf(ctx, "index search content: %v", err)
				return
			}
		default:
			logrus.Warningf("unhandled process step: %v, skipping", step.Action)
		}
	}
}

func (fp *fileProcessor) runRules(ctx context.Context, trigger models.RuleTrigger) error {
	if fp.document == nil {
		return errors.New("no document set")
	}

	rules, err := fp.db.RuleStore.GetActiveUserRules(fp.document.UserId, trigger)
	if err != nil {
		return fmt.Errorf("load rules: %v", err)
	}

	process := &models.ProcessItem{
		DocumentId: fp.document.Id,
		Action:     models.ProcessRules,
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

	log.Context(ctx).
		WithField("user", fp.document.UserId).
		WithField("documentId", fp.document.Id).
		WithField("total-rules", len(rules)).
		WithField("trigger", trigger).
		Infof("Run user rules for document")

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
