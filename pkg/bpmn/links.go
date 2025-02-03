package bpmn

import (
	"fmt"
	"strings"

	"github.com/pbinitiative/zenbpm/pkg/bpmn/model/bpmn20"
	"github.com/pbinitiative/zenbpm/pkg/bpmn/var_holder"
)

func (state *BpmnEngineState) handleIntermediateThrowEvent(process *ProcessInfo, instance *processInstanceInfo, ite bpmn20.TIntermediateThrowEvent, activity activity) (nextCommands []command) {
	linkName := ite.LinkEventDefinition.Name
	if len(strings.TrimSpace(linkName)) == 0 {
		nextCommands = []command{errorCommand{
			err:         newEngineErrorf("missing link name in link intermediate throw event element id=%s name=%s", ite.Id, ite.Name),
			elementId:   ite.Id,
			elementName: ite.Name,
		}}
	}
	for _, ice := range process.definitions.Process.IntermediateCatchEvent {
		if ice.LinkEventDefinition.Name == linkName {
			elementVarHolder := var_holder.New(&instance.VariableHolder, nil)
			if err := propagateProcessInstanceVariables(&elementVarHolder, ite.Output); err != nil {
				msg := fmt.Sprintf("Can't evaluate expression in element id=%s name=%s", ite.Id, ite.Name)
				nextCommands = []command{errorCommand{
					err:         &ExpressionEvaluationError{Msg: msg, Err: err},
					elementId:   ite.Id,
					elementName: ite.Name,
				}}
			} else {
				var element bpmn20.BaseElement = ice
				nextCommands = []command{activityCommand{
					sourceId:       ice.Id,
					element:        &element,
					originActivity: activity,
				}}
			}
			break
		}
	}
	if len(nextCommands) == 0 {
		nextCommands = []command{errorCommand{
			err:         newEngineErrorf("missing link intermediate catch event with linkName=%s", linkName),
			elementId:   ite.Id,
			elementName: ite.Name,
		}}
	}
	return nextCommands
}
