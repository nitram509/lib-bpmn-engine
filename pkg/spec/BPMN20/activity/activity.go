package activity

type LifecycleState string

const (
	Inactive     LifecycleState = "INACTIVE"
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
	Closed       LifecycleState = "CLOSED"
)
