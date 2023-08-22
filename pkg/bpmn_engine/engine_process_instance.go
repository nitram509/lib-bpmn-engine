package bpmn_engine

import (
	"github.com/nitram509/lib-bpmn-engine/pkg/bpmn_engine/var_holder"
	"github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"
	"time"
)

// FIXME: shall this be exported?
type processInstanceInfo struct {
	ProcessInfo    *ProcessInfo              `json:"-"`
	InstanceKey    int64                     `json:"ik"`
	VariableHolder var_holder.VariableHolder `json:"vh"`
	CreatedAt      time.Time                 `json:"c"`
	State          ActivityState             `json:"s"`
	CaughtEvents   []catchEvent              `json:"ce"`
	commandQueue   []*execCommand
}

type ProcessInstance interface {
	GetProcessInfo() *ProcessInfo
	GetInstanceKey() int64

	// GetVariable from the process instance's variable context
	GetVariable(key string) interface{}

	// SetVariable to the process instance's variable context
	SetVariable(key string, value interface{})

	GetCreatedAt() time.Time
	GetState() ActivityState
}

type execCommand struct {
	flowId        string
	source        Activity
	destination   string
	inboundFlowId string
	baseElement   BPMN20.BaseElement
}

func (pii *processInstanceInfo) GetProcessInfo() *ProcessInfo {
	return pii.ProcessInfo
}

func (pii *processInstanceInfo) GetInstanceKey() int64 {
	return pii.InstanceKey
}

func (pii *processInstanceInfo) GetVariable(key string) interface{} {
	return pii.VariableHolder.GetVariable(key)
}

func (pii *processInstanceInfo) SetVariable(key string, value interface{}) {
	pii.VariableHolder.SetVariable(key, value)
}

func (pii *processInstanceInfo) GetCreatedAt() time.Time {
	return pii.CreatedAt
}

// GetState returns one of [ Ready, Active, Completed, Failed ]
func (pii *processInstanceInfo) GetState() ActivityState {
	return pii.State
}

// popCommand return next Command AND deletes from the FIFO queue, if available, else nil
func (pii *processInstanceInfo) popCommand() (cmd *execCommand) {
	if pii.hasCommands() {
		cmd = pii.commandQueue[0]
		pii.commandQueue = pii.commandQueue[1:]
	}
	return cmd
}

// peekCommand return next Command, if available, else nil
func (pii *processInstanceInfo) peekCommand() (cmd *execCommand) {
	if pii.hasCommands() {
		cmd = pii.commandQueue[0]
	}
	return cmd
}

func (pii *processInstanceInfo) hasCommands() bool {
	return len(pii.commandQueue) > 0
}

// appendCommand to the FIFO queue
func (pii *processInstanceInfo) appendCommand(cmd *execCommand) {
	pii.commandQueue = append(pii.commandQueue, cmd)
}

type flowTransition interface {
	Id() string
	Source() Activity
	Destination() Activity
}

func (cmd *execCommand) Id() string {
	return cmd.flowId
}
func (cmd *execCommand) Source() Activity {
	return cmd.source
}
func (cmd *execCommand) Destination() string {
	return cmd.destination
}
