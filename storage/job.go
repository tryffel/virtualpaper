package storage

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type JobStore struct {
	db *sqlx.DB
}

func (s JobStore) Name() string {
	return "Jobs"
}

func (s JobStore) parseError(err error, action string) error {
	return getDatabaseError(err, s, action)
}

// GetByDocument returns all jobs related to document
func (s *JobStore) GetByDocument(documentId string) (*[]models.Job, error) {

	sql := `
SELECT
*
FROM jobs
WHERE jobs.document_id = $1;
`

	jobs := &[]models.Job{}

	err := s.db.Select(jobs, sql, documentId)
	return jobs, s.parseError(err, "get by document")
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
	return jobs, s.parseError(err, "get by user")
}

func (s *JobStore) Create(documentId string, job *models.Job) error {

	sql := `
INSERT INTO jobs (document_id, status, message, started_at, stopped_at)
VALUES ($1, $2, $3, $4, $5) RETURNING id;
`

	val, _ := job.Status.Value()

	res, err := s.db.Query(sql, documentId, val, job.Message, job.StartedAt, job.StoppedAt)
	if err != nil {
		return s.parseError(err, "create")
	}
	defer res.Close()

	if res.Next() {
		var id int
		err := res.Scan(&id)
		if err != nil {
			return s.parseError(err, "create, scan results")
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
	return s.parseError(err, "update")
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
limit 50;
`

	dto := &[]models.ProcessItem{}

	err := s.db.Select(dto, sql)
	if err != nil {
		return dto, 0, s.parseError(err, "get pending processItems")
	}

	sql = `
SELECT COUNT(DISTINCT(document_id, step)) AS count
FROM process_queue;
`
	var n int

	row := s.db.QueryRow(sql)
	err = row.Scan(&n)
	return dto, n, s.parseError(err, "get pending ProcessItems, scan")
}

// GetPendingProcessing returns max 100 documents ordered by process_queue created_at.
// Also returns total number of pending process_queues.
func (s *JobStore) GetDocumentsPendingProcessing() (*[]models.Document, error) {

	sql := `
SELECT d.id AS id, d.user_id AS user_id, d.filename AS filename, d.hash AS hash, d.size AS SIZE, d.date AS DATE
FROM documents d
LEFT JOIN process_queue pq ON d.id = pq.document_id
WHERE pq.step IS NOT NULL
AND pq.running = FALSE
ORDER by pq.created_at ASC
LIMIT 40;
`
	dto := &[]models.Document{}

	err := s.db.Select(dto, sql)
	if err != nil {
		return dto, s.parseError(err, "get documents pending")
	}
	return dto, s.parseError(err, "scan documents pending")
}

// GetDocumentPendingSteps returns ProcessItems not yet started on given document in ascending order.
func (s *JobStore) GetDocumentPendingSteps(documentId string) (*[]models.ProcessItem, error) {
	sql := `
SELECT document_id, step
	FROM process_queue
WHERE running=FALSE
AND document_id = $1
ORDER BY step ASC;
`

	dto := &[]models.ProcessItem{}
	err := s.db.Select(dto, sql, documentId)
	return dto, s.parseError(err, "get pending ProcessSteps")
}

// GetDocumentStatus returns status for given document:
// pending, indexing, ready
func (s *JobStore) GetDocumentStatus(documentId string) (string, error) {
	sql := `
SELECT running
FROM process_queue
WHERE document_id=$1
GROUP BY running;
`

	rows, err := s.db.Query(sql, documentId)
	if err != nil {
		dbERr := s.parseError(err, "get document status for processSteps")
		if errors.Is(dbERr, errors.ErrRecordNotFound) {
			return "ready", nil
		}

		return "", s.parseError(err, "get document status for processSteps")
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
		return nil, s.parseError(err, "start ProcessItem")
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
	return job, s.parseError(err, "mark ProcessStep started")
}

// CreateProcessItem add single process item.
func (s *JobStore) CreateProcessItem(item *models.ProcessItem) error {
	sql := `
INSERT INTO process_queue (document_id, step)
VALUES ($1, $2);
`
	_, err := s.db.Exec(sql, item.DocumentId, item.Step)
	return s.parseError(err, "create ProcessSteps")
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
	return s.parseError(err, "mark ProcessSteps done")
}

// AddDocument adds default processing steps for document. Document must be existing.
func (s *JobStore) AddDocument(doc *models.Document) error {
	return s.addDocument(doc.Id, models.ProcessHash)
}

func (s *JobStore) addDocument(documentId string, fromStep models.ProcessStep) error {
	sql := `
INSERT INTO process_queue (document_id, step)
VALUES 
`

	logrus.Debugf("add document %s for processing starting from step %s", documentId, fromStep)
	var err error
	args := make([]interface{}, len(models.ProcessStepsAll)*2)
	for i := 0; i < len(models.ProcessStepsAll); i++ {
		if i > 0 {
			sql += ", "
		}
		args[i*2] = documentId
		args[i*2+1], err = models.ProcessStepsAll[i].Value()
		if err != nil {
			return fmt.Errorf("insert processStep %s: %v", models.ProcessStepsAll[i], err)
		}

		sql += fmt.Sprintf(" ($%d, $%d)", i*2+1, i*2+2)
	}

	sql += ";"

	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "add document ProcessSteps")

}

// CancelRunningProcesses marks all processes that are currently running as not running.
func (s *JobStore) CancelRunningProcesses() error {
	sql := `
UPDATE process_queue
SET running=FALSE
WHERE running=TRUE
`

	_, err := s.db.Exec(sql)
	return s.parseError(err, "cancel running ProcessItem")
}

// ForceProcessing adds documents to process queue. If documentID != 0, mark only given document. If
// userId != 0, mark all documents for user. Else mark all documents for re-processing. FromStep
// is the first step and successive steps are expected to re-run as well.
func (s *JobStore) ForceProcessing(userId int, documentId string, fromStep models.ProcessStep) error {
	var args []interface{}
	steps := models.ProcessStepsAll[fromStep-1:]
	stepsSql := ""
	for i, v := range steps {
		if i != 0 {
			stepsSql += ", "
		}
		val, _ := v.Value()
		stepsSql += fmt.Sprintf("(%d)", val)
	}

	sql := `
INSERT INTO process_queue (document_id, step)
SELECT documents.id AS document_id, steps.step
FROM documents
JOIN (SELECT DISTINCT * FROM (VALUES %s) AS v) AS steps(step) ON TRUE
`
	sql = fmt.Sprintf(sql, stepsSql)
	if documentId != "" {
		sql += fmt.Sprintf(" WHERE documents.id = $%d", len(args)+1)
		args = append(args, documentId)
	} else if userId != 0 {
		sql += fmt.Sprintf(" WHERE documents.user_id = $%d", len(args)+1)
		args = append(args, userId)
	}

	_, err := s.db.Exec(sql, args...)
	return s.parseError(err, "force processing ProcessSteps")
}

// CancelDocumentProcessing removes all steps from processing queue for document.
func (s *JobStore) CancelDocumentProcessing(documentId string) error {
	sql := `
	DELETE FROM process_queue
	WHERE document_id = $1
`
	_, err := s.db.Exec(sql, documentId)
	return s.parseError(err, "clear document queue")
}
