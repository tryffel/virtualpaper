package models

import (
	"database/sql/driver"
	"fmt"
	"time"
)

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
type Job struct {
	*Timestamp
	Id         int       `db:"id" json:"id"`
	DocumentId int       `db:"document_id" json:"document_id"`
	Message    string    `db:"message" json:"message"`
	Status     JobStatus `json:"status"`

	StartedAt time.Time `db:"started_at" json:"started_at"`
	StoppedAt time.Time `db:"stopped_at" json:"stopped_at"`
}

// JobComposite contains additional information. Actual underlying model is still Job.
type JobComposite struct {
	*Job
	Duration time.Duration `db:"duration" json:"duration"`
}
