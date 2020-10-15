package storage

import (
	"github.com/jmoiron/sqlx"
	"tryffel.net/go/virtualpaper/models"
)

type JobStore struct {
	db *sqlx.DB
}

// GetByDocument returns all jobs related to document
func (s *JobStore) GetByDocument(documentId int) (*[]models.Job, error) {

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

// GetByDocument returns all jobs related to document
func (s *JobStore) GetByUser(userId int, paging Paging) (*[]models.JobComposite, error) {

	sql := `
SELECT
       jobs.id as id,
       jobs.document_id as document_id,
       jobs.message as message,
       jobs.status as status,
       jobs.started_at as started_at,
       jobs.stopped_at as stopped_at,
       extract(epoch from  jobs.stopped_at - jobs.started_at)::numeric::integer as duration
FROM jobs
LEFT JOIN documents d ON jobs.document_id = d.id
WHERE d.user_id = $1
ORDER BY document_id, started_at
OFFSET $2
LIMIT $3;
;
`

	jobs := &[]models.JobComposite{}

	err := s.db.Select(jobs, sql, userId, paging.Offset, paging.Limit)
	return jobs, getDatabaseError(err)
}

func (s *JobStore) Create(documentId int, job *models.Job) error {

	sql := `
INSERT INTO jobs (document_id, status, message, started_at, stopped_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
`

	res, err := s.db.Query(sql, documentId, job.Status, job.Message, job.StartedAt, job.StoppedAt)
	if err != nil {
		return getDatabaseError(err)
	}
	defer res.Close()

	if res.Next() {
		var id int
		err := res.Scan(&id)
		if err != nil {
			return getDatabaseError(err)
		}
		job.Id = id
	}

	return nil
}
