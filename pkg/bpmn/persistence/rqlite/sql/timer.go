package rqlite

import "fmt"

type TimerEntity struct {
	ElementID          string
	ElementInstanceKey int64
	ProcessKey         int64
	ProcessInstanceKey int64
	TimerState         int64
	CreatedAt          int64
	DueDate            int64
	Duration           int64
}

const TIMER_TABLE_CREATE = `CREATE TABLE IF NOT EXISTS timer (
	element_id TEXT NOT NULL,
	element_instance_key INTEGER NOT NULL,
	process_key INTEGER NOT NULL,
	process_instance_key INTEGER NOT NULL,
	state INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	due_at INTEGER NOT NULL,
	duration INTEGER NOT NULL,
	FOREIGN KEY(process_instance_key) REFERENCES process_instance(key)
	FOREIGN KEY(process_key) REFERENCES process_definition(key)
	);`

const TIMER_SELECT = `SELECT * FROM timer WHERE %s;`

const TIMER_INSERT = `INSERT INTO timer
(element_id, element_instance_key, process_key, process_instance_key, state, created_at, due_at, duration)
VALUES
('%s', %d, %d, %d, %d, %d, %d, %d) ON CONFLICT DO UPDATE SET state = %d;`

func BuildTimerUpsertQuery(timer *TimerEntity) string {
	// FIXME: for speed this needs to be splited
	return fmt.Sprintf(TIMER_INSERT, timer.ElementID, timer.ElementInstanceKey, timer.ProcessKey, timer.ProcessInstanceKey, timer.TimerState, timer.CreatedAt, timer.DueDate, timer.Duration, timer.TimerState)

}
