package main

import (
	"encoding/xml"
	"fmt"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
	"io/ioutil"
	"testing"
)

func TestXxx(*testing.T) {
	xmldata, err := ioutil.ReadFile("example1.xml") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var BpmnDefintions BPMN20.TDefinitions
	err = xml.Unmarshal(xmldata, &BpmnDefintions)
	fmt.Println(err)
}
