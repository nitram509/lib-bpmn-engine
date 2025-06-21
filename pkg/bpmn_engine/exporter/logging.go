package exporter

import "fmt"

// LoggingEventExported writes all events to a log file
type LoggingEventExported struct {
}

// NewEventLogExporter creates a new instance of a LoggingEventExported
func NewEventLogExporter() *LoggingEventExported {
	return &LoggingEventExported{}
}

func (*LoggingEventExported) NewProcessEvent(event *ProcessEvent) {
	fmt.Printf("New Process event version: %d, processKey: %d, processID: %s\n", event.Version, event.ProcessKey, event.ProcessId)
}

func (*LoggingEventExported) EndProcessEvent(event *ProcessInstanceEvent) {
	fmt.Printf("End Process event version: %d, processKey: %d, processID: %s, processInstanceKey: %d\n", event.Version, event.ProcessKey, event.ProcessId, event.ProcessInstanceKey)
}

func (*LoggingEventExported) NewProcessInstanceEvent(event *ProcessInstanceEvent) {
	fmt.Printf("New Process Instance version: %d, processKey: %d, processID: %s, processInstanceKey: %d\n", event.Version, event.ProcessKey, event.ProcessId, event.ProcessInstanceKey)
}

func (*LoggingEventExported) NewElementEvent(event *ProcessInstanceEvent, elementInfo *ElementInfo) {
	fmt.Printf("New Element event version: %d, processKey: %d, processID: %s, processInstanceKey: %d, elementType: %s, elementId: %s, intent: %s\n",
		event.Version, event.ProcessKey, event.ProcessId, event.ProcessInstanceKey,
		elementInfo.BpmnElementType, elementInfo.ElementId, elementInfo.Intent)
}
