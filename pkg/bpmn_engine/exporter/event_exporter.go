package exporter

type EventExporter interface {
	NewProcess(eventId int64, processId string, processKey int64, version int32, xmlData []byte, resourceName string, checksum string)
	NewProcessInstance(eventId int64, id string, key2 int64, version int32)
}
