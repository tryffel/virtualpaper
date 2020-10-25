package storage

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
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
		CASE WHEN EXTRACT(epoch FROM jobs.stopped_at) = 0
        	THEN now()
        	ELSE stopped_at
    	END AS stopped_at
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
	if err == nil {
		for i, v := range *jobs {
			v.SetDuration()
			(*jobs)[i].Duration = v.Duration
		}
	}
	return jobs, getDatabaseError(err)
}

func (s *JobStore) Create(documentId int, job *models.Job) error {

	sql := `
INSERT INTO jobs (document_id, status, message, started_at, stopped_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
`

	val, _ := job.Status.Value()

	res, err := s.db.Query(sql, documentId, val, job.Message, job.StartedAt, job.StoppedAt)
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

func (s *JobStore) Update(job *models.Job) error {

	sql := `
UPDATE jobs
SET document_id=$2, status=$3, message=$4, started_at=$5, stopped_at=$6
WHERE id = $1;
`

	val, _ := job.Status.Value()
	_, err := s.db.Exec(sql, job.Id, job.DocumentId, val, job.Message, job.StartedAt, job.StoppedAt)
	return getDatabaseError(err)
}

// GetPendingProcessing returns max 100 processQueue items ordered by created_at.
// Also returns total number of pending process_queues.
func (s *JobStore) GetPendingProcessing() (*[]models.ProcessItem, int, error) {

	sql := `
select d.document_id as document_id, min(d.step) as step, min(d.created_at) as created_at
from (
         select document_id,
                step,
                created_at
         from process_queue
         where running = false
         order by created_at asc
     ) as d
group by d.document_id
order by created_at
limit 20;
`

	dto := &[]models.ProcessItem{}

	err := s.db.Select(dto, sql)
	if err != nil {
		return dto, 0, getDatabaseError(err)
	}

	sql = `
SELECT COUNT(DISTINCT(document_id, step)) AS count
FROM process_queue;
`
	var n int

	row := s.db.QueryRow(sql)
	err = row.Scan(&n)
	return dto, n, getDatabaseError(err)
}

// GetDocumentPendingSteps returns ProcessItems not yet started on given document in ascending order.
func (s *JobStore) GetDocumentPendingSteps(documentId int) (*[]models.ProcessItem, error) {
	sql := `
SELECT document_id, step
	FROM process_queue
WHERE running=FALSE
AND document_id = $1
ORDER BY step ASC;
`

	dto := &[]models.ProcessItem{}
	err := s.db.Select(dto, sql, documentId)
	return dto, getDatabaseError(err)
}

// GetDocumentStatus returns status for given document:
// pending, indexing, ready
func (s *JobStore) GetDocumentStatus(documentId int) (string, error) {
	sql := `
SELECT running
FROM process_queue
WHERE document_id=$1
GROUP BY running;
`

	rows, err := s.db.Query(sql, documentId)
	if err != nil {
		dbERr := getDatabaseError(err)
		if errors.Is(dbERr, ErrRecordNotFound) {
			return "ready", nil
		}

		return "", getDatabaseError(err)
	}

	jobPending := false
	jobRunning := false

	for rows.Next() {
		var val bool
		err = rows.Scan(&val)
		if err != nil {
			logrus.Warningf("unexpected token while 'getDocumentStatus' query: %v", err)
		} else {
			if val {
				jobRunning = true
			} else {
				jobPending = true
			}
		}
	}

	if jobRunning {
		return "indexing", nil
	}

	if jobPending {
		return "pending", nil
	}
	return "ready", nil
}

// StartProcessItem attempts to mark processItem as running. If successful, create corresponding Job and
// return it.
func (s *JobStore) StartProcessItem(item *models.ProcessItem, msg string) (*models.Job, error) {
	sql := `
UPDATE process_queue
SET running=TRUE 
WHERE document_id = $1
AND step = $2
AND running=FALSE;
`
	res, err := s.db.Exec(sql, item.DocumentId, item.Step)
	if err != nil {
		return nil, getDatabaseError(err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		logrus.Errorf("get rows affected: %v", err)
	} else if affected == 0 {
		return nil, errors.New("process item does not exist")
	}

	job := &models.Job{
		Id:         0,
		DocumentId: item.DocumentId,
		Message:    msg,
		Status:     models.JobRunning,
		Step:       item.Step,
		StartedAt:  time.Now(),
		StoppedAt:  time.Time{},
	}

	err = s.Create(item.DocumentId, job)
	return job, getDatabaseError(err)
}

// CreateProcessItem add single process item.
func (s *JobStore) CreateProcessItem(item *models.ProcessItem) error {
	sql := `
INSERT INTO process_queue (document_id, step)
VALUES ($1, $2);
`
	_, err := s.db.Exec(sql, item.DocumentId, item.Step)
	return getDatabaseError(err)
}

// MarkProcessinDone update given item. If ok, remove record, else mark it as not running
func (s *JobStore) MarkProcessingDone(item *models.ProcessItem, ok bool) error {
	var sql string
	if ok {
		sql = `
			DELETE FROM process_queue
			WHERE document_id = $1
			AND step = $2
			AND running =TRUE;
			`
	} else {
		sql = `
			UPDATE process_queue
			SET running=FALSE
			WHERE document_id = $1
			AND step = $2
			AND running=FALSE;
			`
	}
	_, err := s.db.Exec(sql, item.DocumentId, item.Step)
	return getDatabaseError(err)
}

// AddDocument adds default processing steps for document. Document must be existing.
func (s *JobStore) AddDocument(doc *models.Document) error {
	sql := `
INSERT INTO process_queue (document_id, step)
VALUES 
`

	var err error
	args := make([]interface{}, len(models.ProcessStepsAll)*2)
	for i := 0; i < len(models.ProcessStepsAll); i++ {
		if i > 0 {
			sql += ", "
		}
		args[i*2] = doc.Id
		args[i*2+1], err = models.ProcessStepsAll[i].Value()
		if err != nil {
			return fmt.Errorf("insert processStep %s: %v", models.ProcessStepsAll[i], err)
		}

		sql += fmt.Sprintf(" ($%d, $%d)", i*2+1, i*2+2)
	}

	sql += ";"

	_, err = s.db.Exec(sql, args...)
	return getDatabaseError(err)
}

// CancelRunningProcesses marks all processes that are currently running as not running.
func (s *JobStore) CancelRunningProcesses() error {
	sql := `
UPDATE process_queue
SET running=FALSE
WHERE running=TRUE
`

	_, err := s.db.Exec(sql)
	return getDatabaseError(err)
}
