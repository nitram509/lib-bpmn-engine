package rqlite

import (
	"fmt"
)

type ProcessDefinitionEntity struct {
	Key              int64  `json:"key,string"`
	Version          int32  `json:"version"`
	BpmnProcessId    string `json:"bpmnProcessId"`
	BpmnData         string `json:"bpmnData"`
	BpmnChecksum     []byte `json:"bpmnChecksum"`
	BpmnResourceName string `json:"bpmnResourceName"`
}

// Represents a process definition loaded to ProcessInfo
const PROCESS_DEFINITION_TABLE_CREATE = `
CREATE TABLE IF NOT EXISTS process_definition (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	version INTEGER NOT NULL,
	bpmn_process_id TEXT NOT NULL,
	bpmn_data TEXT NOT NULL,
	bpmn_checksum BLOB NOT NULL,
	bpmn_resource_name TEXT NOT NULL
);`

const PROCESS_DEFINITION_INSERT = `
INSERT INTO process_definition
(key, version, bpmn_process_id, bpmn_data, bpmn_checksum, bpmn_resource_name)
VALUES
(%d, %d, '%s', '%s', '%s', '%s');`

// func BuildCreateDefinitionQuery(processDefinition *model.ProcessInfo) string {
// 	return fmt.Sprintf(PROCESS_DEFINITION_INSERT,
// 		processDefinition.Version, processDefinition.BpmnProcessId, base64.StdEncoding.EncodeToString([]byte(processDefinition.BpmnData)), base64.StdEncoding.EncodeToString(processDefinition.BpmnChecksum[:]), processDefinition.BpmnResourceName)
// }

const PROCESS_DEFINITION = `SELECT key, version, bpmn_process_id, bpmn_data, bpmn_checksum, bpmn_resource_name FROM process_definition WHERE %s ORDER BY version DESC;`

func BuildProcessDefinitionUpsertQuery(processDefinition *ProcessDefinitionEntity) string {
	// FIXME: for speed this needs to be splited
	processDefinitionQuery := fmt.Sprintf(PROCESS_DEFINITION_INSERT, processDefinition.Key,
		processDefinition.Version, processDefinition.BpmnProcessId, processDefinition.BpmnData, processDefinition.BpmnChecksum, processDefinition.BpmnResourceName)

	return processDefinitionQuery
}
