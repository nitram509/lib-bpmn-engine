package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/exporter"

// AddEventExporter registers an EventExporter instance
func (state *BpmnEngineState) AddEventExporter(exporter exporter.EventExporter) {
	state.exporters = append(state.exporters, exporter)
}

func (state *BpmnEngineState) publishProcessEvent(processInfo ProcessInfo, xmlData []byte, resourceName string, checksum string) {
	for _, exporter := range state.exporters {
		exporter.NewProcess(
			state.generateKey(),
			processInfo.BpmnProcessId,
			processInfo.ProcessKey,
			processInfo.Version,
			xmlData,
			resourceName,
			checksum,
		)
	}
}

func (state *BpmnEngineState) publishProcessInstanceEvent(processInfo ProcessInfo, xmlData []byte, resourceName string, checksum string) {
	for _, exporter := range state.exporters {
		exporter.NewProcessInstance(
			state.generateKey(),
			processInfo.BpmnProcessId,
			processInfo.ProcessKey,
			processInfo.Version,
		)
	}
}
