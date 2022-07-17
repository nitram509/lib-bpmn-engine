package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"

// AddEventExporter registers an EventExporter instance
func (state *BpmnEngineState) AddEventExporter(exporter exporter.EventExporter) {
	state.exporters = append(state.exporters, exporter)
}

func (state *BpmnEngineState) exportProcessEvent(processInfo ProcessInfo, xmlData []byte, resourceName string, checksum string) {
	event := exporter.ProcessEvent{
		ProcessId:    processInfo.BpmnProcessId,
		ProcessKey:   processInfo.ProcessKey,
		Version:      processInfo.Version,
		XmlData:      xmlData,
		ResourceName: resourceName,
		Checksum:     checksum,
	}
	for _, exporter := range state.exporters {
		exporter.NewProcess(&event)
	}
}

func (state *BpmnEngineState) exportProcessInstanceEvent(process ProcessInfo, processInstance ProcessInstanceInfo) {
	event := exporter.ProcessInstanceEvent{
		ProcessId:          process.BpmnProcessId,
		ProcessKey:         process.ProcessKey,
		Version:            process.Version,
		ProcessInstanceKey: processInstance.instanceKey,
	}
	for _, exporter := range state.exporters {
		exporter.NewProcessInstance(&event)
	}
}
