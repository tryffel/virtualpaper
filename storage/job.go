package storage

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
	"tryffel.net/go/virtualpaper/errors"
	"tryffel.net/go/virtualpaper/models"
)

type JobStore struct {
	db *sqlx.DB
	sq squirrel.StatementBuilderType
}

func (s JobStore) Name() string {
	return "Jobs"
}

func newJobStore(db *sqlx.DB) *JobStore {
	return &JobStore{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

}

func (s JobStore) parseError(err error, action string) error {
	return getDatabaseError(err, s, action)
}

// GetJobsByDocumentId returns all jobs related to document
func (s *JobStore) GetJobsByDocumentId(documentId string) (*[]models.Job, error) {
	sql := `SELECT * FROM jobs WHERE jobs.document_id = $1`

	jobs := &[]models.Job{}
	err := s.db.Select(jobs, sql, documentId)
	return jobs, s.parseError(err, "get by document")
}

// GetJobsByDocumentId returns all jobs related to document
func (s *JobStore) GetJobsByUserId(userId int, paging Paging) (*[]models.JobComposite, error) {
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
LIMIT $3;`

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

func (s *JobStore) CreateJob(documentId string, job *models.Job) error {
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
		res.Close()
	}
	return nil
}

func (s *JobStore) UpdateJob(job *models.Job) error {
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
// Only returns steps for documents that are not currently being processed
func (s *JobStore) GetPendingProcessing() (*[]models.ProcessItem, int, error) {

	sql := `
select 
	d.document_id as document_id, 
	d.action as action,
	min(d.created_at) as created_at,
	d.trigger as trigger
from (
	select document_id, action, created_at, action_order, trigger
	from process_queue
	where running = false 
	-- ignore document's that are being processed
	and document_id not in 
		(
			select document_id 
			from process_queue 
			where running=true group by document_id
		) order by created_at asc
     ) as d
group by d.document_id, action, action_order, d.trigger
order by action_order, created_at
limit 50;
`

	dto := &[]models.ProcessItem{}

	err := s.db.Select(dto, sql)
	if err != nil {
		return dto, 0, s.parseError(err, "get pending processItems")
	}

	sql = `
SELECT COUNT(DISTINCT(document_id, action)) AS count
FROM process_queue;
`
	var n int

	row := s.db.QueryRow(sql)
	err = row.Scan(&n)
	return dto, n, s.parseError(err, "get pending ProcessItems, scan")
}

// GetDocumentsPendingProcessing returns list of document ids that are not currently being processed
// and have processing queued. Only first 50 documents are returned
func (s *JobStore) GetDocumentsPendingProcessing() (*[]string, error) {
	sql := `SELECT document_id FROM process_queue
WHERE document_id NOT IN (
    SELECT document_id FROM process_queue
    WHERE running = true GROUP BY document_id
) 
GROUP BY document_id 
ORDER BY min(created_at) ASC 
LIMIT 50`

	ids := &[]string{}
	err := s.db.Select(ids, sql)
	return ids, s.parseError(err, "get documents pending processing")
}

// GetNextStepForDocument returns next step that hasn't been started yet.
func (s *JobStore) GetNextStepForDocument(documentId string) (*models.ProcessItem, error) {
	sql := `SELECT document_id, action, created_at, trigger FROM process_queue
WHERE document_id = $1 AND running = FALSE
ORDER BY action_order ASC
LIMIT 1`

	step := &models.ProcessItem{}
	err := s.db.Get(step, sql, documentId)
	return step, s.parseError(err, "get next step for document")
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
	rows.Close()

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
UPDATE process_queue SET running=TRUE 
WHERE document_id = $1 AND action = $2 AND running=FALSE;
`
	res, err := s.db.Exec(sql, item.DocumentId, item.Action)
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
		Step:       item.Action,
		StartedAt:  time.Now(),
		StoppedAt:  time.Time{},
	}

	err = s.CreateJob(item.DocumentId, job)
	return job, s.parseError(err, "mark ProcessStep started")
}

// MarkProcessinDone update given item. If ok, remove record, else mark it as not running
func (s *JobStore) MarkProcessingDone(item *models.ProcessItem, ok bool) error {
	var sql string
	if ok {
		sql = `
			DELETE FROM process_queue
			WHERE document_id = $1
			AND action = $2
			AND running =TRUE;
			`
	} else {
		sql = `
			UPDATE process_queue
			SET running=FALSE
			WHERE document_id = $1
			AND action = $2
			AND running=FALSE;
			`
	}
	_, err := s.db.Exec(sql, item.DocumentId, item.Action)
	return s.parseError(err, "mark ProcessSteps done")
}

// ProcessDocumentAllSteps adds default processing steps for document. Document must be existing.
func (s *JobStore) ProcessDocumentAllSteps(documentId string, trigger models.RuleTrigger) error {
	sq := s.sq.Insert("process_queue").Columns("document_id", "action", "action_order", "trigger")
	for _, v := range models.ProcessStepsAll {
		sq = sq.Values(documentId, v, models.ProcessStepsOrder[v], trigger)
	}
	sql, args, err := sq.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}

	logrus.Debugf("add document %s for processing starting from step %s", documentId, models.ProcessHash)
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

// ForceProcessingDocument adds documents to process queue. If documentID != 0, mark only given document. If
// userId != 0, mark all documents for user. Else mark all documents for re-processing. FromStep
// is the first step and successive steps are expected to re-run as well.
func (s *JobStore) ForceProcessingDocument(documentId string, steps []models.ProcessStep) error {
	sq := s.sq.Insert("process_queue").Columns("document_id", "action", "action_order", "trigger")
	for _, v := range steps {
		sq = sq.Values(documentId, v, models.ProcessStepsOrder[v], models.RuleTriggerUpdate)
	}
	sql, args, err := sq.ToSql()
	if err != nil {
		return fmt.Errorf("sql: %v", err)
	}
	_, err = s.db.Exec(sql, args...)
	return s.parseError(err, "force processing ProcessSteps")
}

func (s *JobStore) ForceProcessingByUser(userId int, steps []models.ProcessStep) error {
	stepsSql := ""
	for i, v := range steps {
		if i > 0 {
			stepsSql += ","
		}
		stepsSql += fmt.Sprintf("('%s', %d, '%s')", v, models.ProcessStepsOrder[v], models.RuleTriggerUpdate)
	}

	sql := `INSERT INTO process_queue (document_id, action, action_order, trigger)
	SELECT d.id as document_id, steps.action, steps.action_order
	FROM documents d
	JOIN (
		SELECT * FROM (VALUES %s) AS v (action, action_order)
	) steps ON TRUE`

	var err error
	sql = fmt.Sprintf(sql, stepsSql)
	if userId != 0 {
		sql += " WHERE d.user_id=$1"
		_, err = s.db.Exec(sql, userId)
	} else {
		_, err = s.db.Exec(sql)
	}
	return s.parseError(err, "schedule processing job for documents")
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

// IndexDocumentsByMetadata adds all documents that match the identifiers.
// If user != 0, use has to own the document,
// if keyId != 0, document has to have key,
// if valueId != 0, document has to have the value.
// Either key or value must be supplied.
func (s *JobStore) IndexDocumentsByMetadata(userId int, keyId int, valueId int) error {
	if valueId == 0 && keyId == 0 {
		e := errors.ErrInvalid
		e.ErrMsg = "no key nor value supplied"
	}

	stepSql := fmt.Sprintf("('%s', %d)", models.ProcessFts, models.ProcessStepsOrder[models.ProcessFts])
	selectQuery := s.sq.Select("documents.id as document_id, steps.action, steps.action_order", "'document-update'").
		From("documents").
		LeftJoin("document_metadata dm on documents.id = dm.document_id").
		Join(fmt.Sprintf("(SELECT DISTINCT * FROM (VALUES %s) AS v) AS steps(action, action_order) ON TRUE", stepSql))

	if userId != 0 {
		selectQuery = selectQuery.Where("documents.user_id=?", userId)
	}
	if keyId != 0 {
		selectQuery = selectQuery.Where("dm.key_id=?", keyId)
	}
	if valueId != 0 {
		selectQuery = selectQuery.Where("dm.value_id=?", valueId)
	}
	query := s.sq.Insert("process_queue").
		Columns("document_id", "action", "action_order", "trigger").
		Select(selectQuery)

	sql, args, err := query.ToSql()
	if err != nil {
		e := errors.ErrInternalError
		e.Err = err
		return e
	}

	_, err = s.db.Exec(sql, args...)
	return getDatabaseError(err, s, "queue documents by metadata")
}

func (s *JobStore) AddDocuments(exec SqlExecer, userId int, documents []string, steps []models.ProcessStep, trigger models.RuleTrigger) error {

	stepsSql := ""
	for i, v := range steps {
		if i != 0 {
			stepsSql += ", "
		}
		val, _ := v.Value()
		stepsSql += fmt.Sprintf("('%s', %d)", val, models.ProcessStepsOrder[v])
	}

	selectQuery := s.sq.Select("documents.id as document_id, steps.action, steps.action_order", fmt.Sprintf("'%s'", trigger)).
		From("documents").
		Join(fmt.Sprintf("(SELECT DISTINCT * FROM (VALUES %s) AS v) AS steps(action, action_order) ON TRUE", stepsSql))

	if userId != 0 {
		selectQuery = selectQuery.Where("documents.user_id=?", userId)
	}

	selectQuery = selectQuery.Where(squirrel.Eq{"documents.id": documents})
	query := s.sq.Insert("process_queue").
		Columns("document_id", "action", "action_order").
		Select(selectQuery)

	_, err := exec.ExecSq(query)
	return getDatabaseError(err, s, "queue documents by metadata")
}
