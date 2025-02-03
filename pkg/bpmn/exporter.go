package bpmn

import (
	"github.com/pbinitiative/zenbpm/pkg/bpmn/exporter"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
)

// AddEventExporter registers an EventExporter instance
func (state *BpmnEngineState) AddEventExporter(exporter exporter.EventExporter) {
	state.exporters = append(state.exporters, exporter)
}

func (state *BpmnEngineState) exportNewProcessEvent(processInfo ProcessInfo, xmlData []byte, resourceName string, checksum string) {
	event := exporter.ProcessEvent{
		ProcessId:    processInfo.BpmnProcessId,
		ProcessKey:   processInfo.ProcessKey,
		Version:      processInfo.Version,
		XmlData:      xmlData,
		ResourceName: resourceName,
		Checksum:     checksum,
	}
	for _, exp := range state.exporters {
		exp.NewProcessEvent(&event)
	}
}

func (state *BpmnEngineState) exportEndProcessEvent(process ProcessInfo, processInstance processInstanceInfo) {
	event := exporter.ProcessInstanceEvent{
		ProcessId:          process.BpmnProcessId,
		ProcessKey:         process.ProcessKey,
		Version:            process.Version,
		ProcessInstanceKey: processInstance.InstanceKey,
	}
	for _, exp := range state.exporters {
		exp.EndProcessEvent(&event)
	}
}

func (state *BpmnEngineState) exportProcessInstanceEvent(process ProcessInfo, processInstance processInstanceInfo) {
	event := exporter.ProcessInstanceEvent{
		ProcessId:          process.BpmnProcessId,
		ProcessKey:         process.ProcessKey,
		Version:            process.Version,
		ProcessInstanceKey: processInstance.InstanceKey,
	}
	for _, exp := range state.exporters {
		exp.NewProcessInstanceEvent(&event)
	}
}

func (state *BpmnEngineState) exportElementEvent(process ProcessInfo, processInstance processInstanceInfo, element bpmn20.BaseElement, intent exporter.Intent) {
	event := exporter.ProcessInstanceEvent{
		ProcessId:          process.BpmnProcessId,
		ProcessKey:         process.ProcessKey,
		Version:            process.Version,
		ProcessInstanceKey: processInstance.InstanceKey,
	}
	info := exporter.ElementInfo{
		BpmnElementType: string(element.GetType()),
		ElementId:       element.GetId(),
		Intent:          string(intent),
	}
	for _, exp := range state.exporters {
		exp.NewElementEvent(&event, &info)
	}
}

func (state *BpmnEngineState) exportSequenceFlowEvent(process ProcessInfo, processInstance processInstanceInfo, flow bpmn20.TSequenceFlow) {
	event := exporter.ProcessInstanceEvent{
		ProcessId:          process.BpmnProcessId,
		ProcessKey:         process.ProcessKey,
		Version:            process.Version,
		ProcessInstanceKey: processInstance.InstanceKey,
	}
	info := exporter.ElementInfo{
		BpmnElementType: string(bpmn20.SequenceFlow),
		ElementId:       flow.Id,
		Intent:          string(exporter.SequenceFlowTaken),
	}
	for _, exp := range state.exporters {
		exp.NewElementEvent(&event, &info)
	}
}
