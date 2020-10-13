package models

import "time"

type JobStatus int

const (
	JobAwaiting JobStatus = iota
	JobRunning
	JobFinished
	JobFailure
)

func (j JobStatus) String() string {
	switch j {
	case JobAwaiting:
		return "Waiting"
	case JobRunning:
		return "Running"
	case JobFinished:
		return "Finished"
	case JobFailure:
		return "Failure"
	}
	return "Unknown"
}

// Job is a pipeline that each document goes through. It consists of multiple steps to process document.
type Job struct {
	*Timestamp
	Id         int    `db:"id"`
	DocumentId int    `db:"document_id"`
	Message    string `db:"message"`
	Status     int

	StartedAt time.Time `db:"started_at"`
	StoppedAt time.Time `db:"stopped_at"`
}
