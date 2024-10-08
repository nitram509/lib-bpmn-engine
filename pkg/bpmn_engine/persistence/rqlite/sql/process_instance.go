package rqlite

import "fmt"

type ProcessInstanceEntity struct {
	Key                  int64
	ProcessDefinitionKey int64
	CreatedAt            int64
	State                int
	VariableHolder       string
	CaughtEvents         string
	Activities           string
}

const PROCESS_INSTANCE_TABLE_CREATE = `
CREATE TABLE IF NOT EXISTS process_instance (
	key INTEGER PRIMARY KEY,
	process_definition_key INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	state INTEGER NOT NULL,
	variable_holder TEXT NOT NULL,
	caught_events TEXT NOT NULL,
	activities TEXT NOT NULL,
	FOREIGN KEY(process_definition_key) REFERENCES process_definition(key)
	);`

const PROCESS_INSTANCE_INSERT = `
	INSERT INTO process_instance
	(key, process_definition_key, created_at, state, variable_holder, caught_events, activities)
	VALUES
	(%d, %d, %d, %d, '%s', '%s', '%s') ON CONFLICT DO UPDATE SET state = %d, variable_holder = '%s', caught_events = '%s', activities = '%s';`

const PROCESS_INSTANCE_SELECT = `SELECT key, process_definition_key, created_at, state, variable_holder, caught_events, activities FROM process_instance WHERE %s ORDER BY created_at DESC;`

func BuildProcessInstanceUpsertQuery(pie *ProcessInstanceEntity) string {
	// FIXME: for speed this needs to be splited
	return fmt.Sprintf(PROCESS_INSTANCE_INSERT, pie.Key, pie.ProcessDefinitionKey, pie.CreatedAt, pie.State, pie.VariableHolder, pie.CaughtEvents, pie.Activities, pie.State, pie.VariableHolder, pie.CaughtEvents, pie.Activities)
}
