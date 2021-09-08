package engine

import (
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
)

type Node struct {
	Name string
	Id   string
}

type registeredProcess struct {
	processInfo ProcessInfo
}

type BpmnEngine interface {
	LoadFromDefinitions(definitions BPMN20.TDefinitions)
}

type BpmnEngineCore struct {
	registry []registeredProcess
}

func (core BpmnEngineCore) LoadFromDefinitions(definitions BPMN20.TDefinitions) (DeployResult, error) {
	info := ProcessInfo{
		bpmnProcessId: "123",
		version:       1,
		processKey:    456,
		resourceName:  "xxx",
	}

	core.registry = append(core.registry, registeredProcess{info})

	result := DeployResult{}
	result.key = "1234567890"
	result.processes = append(result.processes, info)
	return result, nil
}

type DeployResult struct {
	key       string
	processes []ProcessInfo
}

type ProcessInfo struct {
	bpmnProcessId string
	version       uint64
	processKey    uint64
	resourceName  string
}
