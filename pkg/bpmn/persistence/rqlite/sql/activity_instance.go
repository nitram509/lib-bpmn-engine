package rqlite

import "fmt"

type ActivityInstanceEntity struct {
	Key                  int64
	ProcessInstanceKey   int64
	ProcessDefinitionKey int64
	CreatedAt            int64
	State                string
	ElementId            string
	BpmnElementType      string
}

const ACTIVITY_INSTANCE_TABLE_CREATE = `
CREATE TABLE IF NOT EXISTS activity_instance (
	key INTEGER PRIMARY KEY,
	process_instance_key INTEGER NOT NULL,
	process_definition_key INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	state TEXT NOT NULL,

	element_id TEXT NOT NULL,
	bpmn_element_type TEXT NOT NULL,

	FOREIGN KEY(process_instance_key) REFERENCES process_instance(key)
	FOREIGN KEY(process_definition_key) REFERENCES process_definition(key)
	);`

const ACTIVITY_INSTANCE_INSERT = `
	INSERT INTO activity_instance
	(key, process_instance_key, process_definition_key, created_at, state, element_id, bpmn_element_type)
	VALUES
	(%d, %d, %d, %d, '%s', '%s', '%s') ;`

const ACTIVITY_INSTANCE_SELECT = `SELECT key, process_instance_key, process_definition_key, created_at, state, element_id, bpmn_element_type FROM activity_instance WHERE %s ORDER BY key ASC;`

func BuildActivityInstanceUpsertQuery(activityInstanceKey int64, processInstanceKey int64, processDefinitionKey int64, createdAt int64, state string, elementId string, bpmnElementType string) string {
	// FIXME: for speed this needs to be splited
	return fmt.Sprintf(ACTIVITY_INSTANCE_INSERT, activityInstanceKey, processInstanceKey, processDefinitionKey, createdAt, state, elementId, bpmnElementType)

}
