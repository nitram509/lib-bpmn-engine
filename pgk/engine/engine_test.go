package engine

import (
	"encoding/xml"
	"fmt"
	"github.com/corbym/gocrest/is"
	"github.com/corbym/gocrest/then"
	"github.com/nitram509/golib-bpmn-model/pgk/spec/BPMN/20100501/BPMN20"
	"io/ioutil"
	"testing"
)

func TestLoadFromBpmnDefinitions(t *testing.T) {

	xmldata, err := ioutil.ReadFile("../../test/simple_task.xml") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	var definitions BPMN20.TDefinitions
	err = xml.Unmarshal(xmldata, &definitions)
	fmt.Println(err)

	core := BpmnEngineCore{}
	deployResult, err := core.LoadFromDefinitions(definitions)

	then.AssertThat(t, err, is.Nil())
	then.AssertThat(t, deployResult.key, is.Not(is.Empty()))

}
