package rqlite

import "fmt"

type MessageSubscriptionEntity struct {
	ElementID           string
	ElementInstanceKey  int64
	ProcessKey          int64
	ProcessInstanceKey  int64
	MessageName         string
	State               int
	CreatedAt           int64
	OriginActivityKey   int64
	OriginActivityState int
	OriginActivityId    string //Bpmn id
}

const MESSAGE_SUBSCRIPTION_TABLE_CREATE = `CREATE TABLE IF NOT EXISTS message_subscription (
	element_instance_key INTEGER PRIMARY KEY,
	element_id TEXT NOT NULL,
	process_key INTEGER NOT NULL,
	process_instance_key INTEGER NOT NULL,
	name TEXT NOT NULL,
	state INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	origin_activity_key INTEGER NOT NULL,
	origin_activity_state INTEGER NOT NULL,
	origin_activity_id TEXT NOT NULL,
	FOREIGN KEY(process_instance_key) REFERENCES process_instance(key)
	FOREIGN KEY(process_key) REFERENCES process_definition(key)
	);`

const MESSAGE_SUBSCRIPTION_SELECT = `SELECT * FROM message_subscription WHERE %s ORDER BY created_at DESC;`

const MESSAGE_SUBSCRIPTION_INSERT = `INSERT INTO message_subscription
(element_instance_key,element_id,  process_key, process_instance_key, name, state, created_at,origin_activity_key, origin_activity_state, origin_activity_id)
VALUES
(%d,'%s',  %d, %d, '%s', %d, %d, %d, %d, '%s') ON CONFLICT DO UPDATE SET state = %d;`

func BuildMessageSubscriptionUpsertQuery(ms *MessageSubscriptionEntity) string {
	// FIXME: for speed this needs to be splited
	return fmt.Sprintf(MESSAGE_SUBSCRIPTION_INSERT, ms.ElementInstanceKey, ms.ElementID, ms.ProcessKey, ms.ProcessInstanceKey, ms.MessageName, ms.State, ms.CreatedAt, ms.OriginActivityKey, ms.OriginActivityState, ms.OriginActivityId, ms.State)
}
