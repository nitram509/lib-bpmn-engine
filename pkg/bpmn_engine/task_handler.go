package bpmn_engine

import "github.com/nitram509/lib-bpmn-engine/pkg/spec/BPMN20"

type taskMatcher func(element *BPMN20.TaskElement) bool

type taskHandlerType string

const (
	taskHandlerForId              = "TASK_HANDLER_ID"
	taskHandlerForType            = "TASK_HANDLER_TYPE"
	taskHandlerForAssignee        = "TASK_HANDLER_ASSIGNEE"
	taskHandlerForCandidateGroups = "TASK_HANDLER_CANDIDATE_GROUPS"
)

type taskHandler struct {
	handlerType taskHandlerType
	matches     taskMatcher
	handler     func(job ActivatedJob)
}

type newTaskHandlerCommand struct {
	handlerType taskHandlerType
	matcher     taskMatcher
	append      func(handler *taskHandler)
}

type NewTaskHandlerCommand2 interface {
	// Handler is the actual handler to be executed
	Handler(func(job ActivatedJob))
}

type NewTaskHandlerCommand1 interface {
	// Id defines a handler for a given element ID (as defined in the task element in the BPMN file)
	// This is 1:1 relation between a handler and a task definition (since IDs are supposed to be unique).
	Id(id string) NewTaskHandlerCommand2

	// Type defines a handler for a Service Task with a given 'type';
	// Hereby 'type' is defined as 'taskDefinition' extension element in the BPMN file.
	// This allows a single handler to be used for multiple task definitions.
	Type(taskType string) NewTaskHandlerCommand2

	// Assignee defines a handler for a User Task with a given 'assignee';
	// Hereby 'assignee' is defined as 'assignmentDefinition' extension element in the BPMN file.
	Assignee(assignee string) NewTaskHandlerCommand2

	// CandidateGroups defines a handler for a User Task with given 'candidate groups';
	// For the handler you can specify one or more groups.
	// If at least one matches a given user task, the handler will be called.
	CandidateGroups(groups ...string) NewTaskHandlerCommand2
}

// NewTaskHandler registers a handler function to be called for service tasks with a given taskId
func (state *BpmnEngineState) NewTaskHandler() NewTaskHandlerCommand1 {
	cmd := newTaskHandlerCommand{
		append: func(handler *taskHandler) {
			state.taskHandlers = append(state.taskHandlers, handler)
		},
	}
	return cmd
}

// Id implements NewTaskHandlerCommand1
func (thc newTaskHandlerCommand) Id(id string) NewTaskHandlerCommand2 {
	thc.matcher = func(element *BPMN20.TaskElement) bool {
		return (*element).GetId() == id
	}
	thc.handlerType = taskHandlerForId
	return thc
}

// Type implements NewTaskHandlerCommand1
func (thc newTaskHandlerCommand) Type(taskType string) NewTaskHandlerCommand2 {
	thc.matcher = func(element *BPMN20.TaskElement) bool {
		return (*element).GetTaskDefinitionType() == taskType
	}
	thc.handlerType = taskHandlerForType
	return thc
}

// Handler implements NewTaskHandlerCommand2
func (thc newTaskHandlerCommand) Handler(f func(job ActivatedJob)) {
	th := taskHandler{
		handlerType: thc.handlerType,
		matches:     thc.matcher,
		handler:     f,
	}
	thc.append(&th)
}

// Assignee implements NewTaskHandlerCommand2
func (thc newTaskHandlerCommand) Assignee(assignee string) NewTaskHandlerCommand2 {
	thc.matcher = func(element *BPMN20.TaskElement) bool {
		return (*element).GetAssignmentAssignee() == assignee
	}
	thc.handlerType = taskHandlerForAssignee
	return thc
}

// CandidateGroups implements NewTaskHandlerCommand2
func (thc newTaskHandlerCommand) CandidateGroups(groups ...string) NewTaskHandlerCommand2 {
	thc.matcher = func(element *BPMN20.TaskElement) bool {
		for _, group := range groups {
			if contains((*element).GetAssignmentCandidateGroups(), group) {
				return true
			}
		}
		return false
	}
	thc.handlerType = taskHandlerForCandidateGroups
	return thc
}

func (state *BpmnEngineState) findTaskHandler(element *BPMN20.TaskElement) func(job ActivatedJob) {
	searchOrder := []taskHandlerType{taskHandlerForId}
	if (*element).GetType() == BPMN20.ServiceTask {
		searchOrder = append(searchOrder, taskHandlerForType)
	}
	if (*element).GetType() == BPMN20.UserTask {
		searchOrder = append(searchOrder, taskHandlerForAssignee, taskHandlerForCandidateGroups)
	}
	for _, handlerType := range searchOrder {
		for _, handler := range state.taskHandlers {
			if handler.handlerType == handlerType {
				if handler.matches(element) {
					return handler.handler
				}
			}
		}
	}
	return nil
}
