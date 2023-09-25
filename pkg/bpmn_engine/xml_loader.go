package bpmn_engine

import (
	"bytes"
	"compress/flate"
	"crypto/md5"
	"encoding/ascii85"
	"encoding/hex"
	"encoding/xml"
	"io"
	"os"

	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
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
		Version:          1,
		BpmnProcessId:    definitions.Process.Id,
		ProcessKey:       state.generateKey(),
		definitions:      definitions,
		bpmnData:         compressAndEncode(xmlData),
		bpmnResourceName: resourceName,
		bpmnChecksum:     md5sum,
	}
	for _, process := range state.processes {
		if process.BpmnProcessId == definitions.Process.Id {
			if areEqual(process.bpmnChecksum, md5sum) {
				return process, nil
			} else {
				processInfo.Version = process.Version + 1
			}
		}
	}
	state.processes = append(state.processes, &processInfo)

	state.exportNewProcessEvent(processInfo, xmlData, resourceName, hex.EncodeToString(md5sum[:]))
	return &processInfo, nil
}

func compressAndEncode(data []byte) string {
	buffer := bytes.Buffer{}
	ascii85Writer := ascii85.NewEncoder(&buffer)
	flateWriter, err := flate.NewWriter(ascii85Writer, flate.BestCompression)
	if err != nil {
		panic(err)
	}
	_, err = flateWriter.Write(data)
	if err != nil {
		panic(err)
	}
	_ = flateWriter.Flush()
	_ = flateWriter.Close()
	_ = ascii85Writer.Close()
	return buffer.String()
}

func decodeAndDecompress(data string) []byte {
	ascii85Reader := ascii85.NewDecoder(bytes.NewBuffer([]byte(data)))
	deflateReader := flate.NewReader(ascii85Reader)
	buffer := bytes.Buffer{}
	_, err := io.Copy(&buffer, deflateReader)
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
