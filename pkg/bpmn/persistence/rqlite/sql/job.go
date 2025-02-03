package rqlite

import "fmt"

type JobEntity struct {
	Key                int64
	ElementID          string
	ElementInstanceKey int64
	ProcessInstanceKey int64
	State              int64
	CreatedAt          int64
}

const JOB_TABLE_CREATE = `CREATE TABLE IF NOT EXISTS job (
	key INTEGER PRIMARY KEY,
	element_id TEXT NOT NULL,
	element_instance_key INTEGER NOT NULL,
	process_instance_key INTEGER NOT NULL,
	state INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	FOREIGN KEY(process_instance_key) REFERENCES process_instance(key)
	);`

const JOB_SELECT = `SELECT * FROM job WHERE %s;`

const JOB_INSERT = `INSERT INTO job
(key, element_id, element_instance_key, process_instance_key, state, created_at)
VALUES
(%d, '%s', %d, %d, %d, %d) ON CONFLICT DO UPDATE SET state = %d;`

func BuildJobUpsertQuery(job *JobEntity) string {
	// FIXME: for speed this needs to be splited
	return fmt.Sprintf(JOB_INSERT, job.Key, job.ElementID, job.ElementInstanceKey, job.ProcessInstanceKey, job.State, job.CreatedAt, job.State)

}
