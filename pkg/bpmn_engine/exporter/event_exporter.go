package exporter

type EventExporter interface {
	NewProcess(event *ProcessEvent)
	NewProcessInstance(event *ProcessInstanceEvent)
}

type ProcessEvent struct {
	ProcessId    string
	ProcessKey   int64
	Version      int32
	XmlData      []byte
	ResourceName string
	Checksum     string
}

type ProcessInstanceEvent struct {
	ProcessId          string
	ProcessKey         int64
	Version            int32
	ProcessInstanceKey int64
}
