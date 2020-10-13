package storage

import (
	"github.com/jmoiron/sqlx"
	"tryffel.net/go/virtualpaper/models"
)

type JobStore struct {
	db *sqlx.DB
}

// GetJobs returns all jobs related to document
func (s *JobStore) GetJobs(documentId int) (*[]models.Job, error) {

	sql := `
SELECT
*
FROM jobs
WHERE jobs.document_id = $1;
`

	jobs := &[]models.Job{}

	err := s.db.Select(jobs, sql, documentId)
	return jobs, getDatabaseError(err)
}
