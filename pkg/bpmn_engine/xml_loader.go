package bpmn_engine

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/xml"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"os"
)

// LoadFromFile loads a given BPMN file by filename into the engine
// and returns ProcessInfo details for the deployed workflow
func (state *BpmnEngineState) LoadFromFile(filename string) (*ProcessInfo, error) {
	xmlData, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return state.load(xmlData, filename)
}

// LoadFromBytes loads a given BPMN file by xmlData byte array into the engine
// and returns ProcessInfo details for the deployed workflow
func (state *BpmnEngineState) LoadFromBytes(xmlData []byte) (*ProcessInfo, error) {
	return state.load(xmlData, "")
}

func (state *BpmnEngineState) load(xmlData []byte, resourceName string) (*ProcessInfo, error) {
	md5sum := md5.Sum(xmlData)
	var definitions BPMN20.TDefinitions
	err := xml.Unmarshal(xmlData, &definitions)
	if err != nil {
		return nil, err
	}

	processInfo := ProcessInfo{
		Version:     1,
		Definitions: definitions,
	}
	for _, process := range state.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.ChecksumBytes, md5sum) {
				return &process, nil
			} else {
				processInfo.Version = process.Version + 1
			}
		}
	}
	processInfo.BpmnProcessId = definitions.Process.Id
	processInfo.ProcessKey = state.generateKey()
	processInfo.ChecksumBytes = md5sum
	state.processes = append(state.processes, processInfo)

	state.exportNewProcessEvent(processInfo, xmlData, resourceName, hex.EncodeToString(md5sum[:]))

	return &processInfo, nil
}
