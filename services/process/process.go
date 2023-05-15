package process

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
	"tryffel.net/go/virtualpaper/storage"
)

func (fp *fileProcessor) completeProcessingStep(process *models.ProcessItem, job *models.Job) {
	fp.Debug("processing completed, status: %v", job.Status)

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
	err = fp.db.JobStore.Update(job)
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

	defer fp.cleanup()

	for _, step := range *pendingSteps {
		fp.Info("run step %s", step.Step.String())

		switch step.Step {
		case models.ProcessHash:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing("file not found")
				if err != nil {
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.updateHash(fp.document)
			if err != nil {
				logrus.Errorf("update hash: %v", err)
				return
			}
		case models.ProcessThumbnail:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing("file not found")
				if err != nil {
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.generateThumbnail()
			if err != nil {
				logrus.Errorf("generate thumbnail: %v", err)
				return
			}
		case models.ProcessParseContent:
			err = fp.ensureFileOpenAndLogFailure()
			if err != nil {
				err = fp.cancelDocumentProcessing("file not found")
				if err != nil {
					logrus.Errorf("cancel document processing: %v", err)
				}
				return
			}
			err := fp.parseContent()
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
