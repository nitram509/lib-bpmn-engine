package BPMN20

type ServiceTask struct {
	Id    string
	scope Scope
}

type Scope struct {
	State LifecycleState
	//dataObjects
	//events
	//conversations
}

type LifecycleState int8

const (
	// todo do we need a 'none'?
	ActivatedState      LifecycleState = 0
	InExecutionState    LifecycleState = 1
	CompletedState      LifecycleState = 2
	InCompensationState LifecycleState = 3
	CompensationState   LifecycleState = 4
	InErrorState        LifecycleState = 5
	InCancellationState LifecycleState = 6
	CancelledState      LifecycleState = 7
)
