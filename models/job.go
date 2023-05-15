package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// JobStatus describes status for process-step
// It maps as integer to database.
type JobStatus string

func (j *JobStatus) Value() (driver.Value, error) {
	switch *j {
	case JobAwaiting:
		return 0, nil
	case JobRunning:
		return 1, nil
	case JobFinished:
		return 2, nil
	case JobFailure:
		return 3, nil
	default:
		return 0, fmt.Errorf("unknown status: %s", *j)
	}
}

func (j *JobStatus) Scan(src interface{}) error {
	var val int
	if valDefault, ok := src.(int); ok {
		val = valDefault
	} else if val32, ok := src.(int32); ok {
		val = int(val32)
	} else if val64, ok := src.(int64); ok {
		val = int(val64)
	} else {
		return fmt.Errorf("expect int, got: %v", src)
	}
	switch val {
	case 0:
		*j = JobAwaiting
	case 1:
		*j = JobRunning
	case 2:
		*j = JobFinished
	case 3:
		*j = JobFailure
	default:
		return fmt.Errorf("unknown status: %d", val)
	}
	return nil
}

const (
	JobAwaiting JobStatus = "Awaiting"
	JobRunning  JobStatus = "Running"
	JobFinished JobStatus = "Finished"
	JobFailure  JobStatus = "Failure"
)

// Job is a pipeline that each document goes through. It consists of multiple steps to process document.
// Job is only used for logging purposes.
type Job struct {
	Id         int         `db:"id" json:"id"`
	DocumentId string      `db:"document_id" json:"document_id"`
	Message    string      `db:"message" json:"message"`
	Status     JobStatus   `json:"status"`
	Step       ProcessStep `db:"process_step"`

	StartedAt time.Time `db:"started_at" json:"started_at"`
	StoppedAt time.Time `db:"stopped_at" json:"stopped_at"`
}

// JobComposite contains additional information. Actual underlying model is still Job.
type JobComposite struct {
	Job
	Duration time.Duration `db:"duration" json:"duration"`
}

func (jc *JobComposite) SetDuration() {
	jc.Duration = jc.StoppedAt.Sub(jc.StartedAt)
}

// ProcessStep describes next step for document.
// It maps as integer to database.
type ProcessStep string

const (
	ProcessHash         ProcessStep = "hash"
	ProcessThumbnail    ProcessStep = "thumbnail"
	ProcessParseContent ProcessStep = "extract"
	ProcessRules        ProcessStep = "rules"
	ProcessFts          ProcessStep = "fts"
)

// ProcessStepsAll is a list of default steps to run for new document.
var ProcessStepsAll = []ProcessStep{ProcessHash, ProcessThumbnail, ProcessParseContent, ProcessRules, ProcessFts}

// ProcessStepsOrder is the order in which the steps are to be run in ascending order.
var ProcessStepsOrder = map[ProcessStep]int{
	ProcessHash:         1,
	ProcessThumbnail:    2,
	ProcessParseContent: 3,
	ProcessRules:        4,
	ProcessFts:          5,
}

var ProcessStepsKeys = map[ProcessStep]string{
	ProcessHash:         "hash",
	ProcessThumbnail:    "thumbnail",
	ProcessParseContent: "content",
	ProcessRules:        "rules",
	ProcessFts:          "fts",
}

func (ps *ProcessStep) Value() (driver.Value, error) {
	return string(*ps), nil
}

func (ps *ProcessStep) Scan(src interface{}) error {
	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("value '%v' not string", src)
	}
	*ps = ProcessStep(str)
	return nil
}

func (ps ProcessStep) String() string {
	return string(ps)
}

// ProcessItem contains document that awaits further processing.
type ProcessItem struct {
	DocumentId string `db:"document_id"`
	Document   *Document
	Action     ProcessStep `db:"action"`
	CreatedAt  time.Time   `db:"created_at"`
}
