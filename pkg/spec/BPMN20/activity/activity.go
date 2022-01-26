package activity

// LifecycleState as per BPMN 2.0 spec, section 13.2.2 Activity, page 428
type LifecycleState string

const (
	Ready        LifecycleState = "READY"
	Active       LifecycleState = "ACTIVE"
	WithDrawn    LifecycleState = "WITHDRAWN"
	Completing   LifecycleState = "COMPLETING"
	Completed    LifecycleState = "COMPLETED"
	Failing      LifecycleState = "FAILING"
	Terminating  LifecycleState = "TERMINATING"
	Compensating LifecycleState = "COMPENSATING"
	Failed       LifecycleState = "FAILED"
	Terminated   LifecycleState = "TERMINATED"
	Compensated  LifecycleState = "COMPENSATED"
)
