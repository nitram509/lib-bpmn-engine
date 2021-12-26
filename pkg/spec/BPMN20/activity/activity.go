package activity

type LifecycleState int8

const (
	Inactive     LifecycleState = 0
	Ready        LifecycleState = 1
	Active       LifecycleState = 2
	WithDrawn    LifecycleState = 3
	Completing   LifecycleState = 4
	Completed    LifecycleState = 5
	Failing      LifecycleState = 6
	Terminating  LifecycleState = 7
	Compensating LifecycleState = 8
	Failed       LifecycleState = 9
	Terminated   LifecycleState = 10
	Compensated  LifecycleState = 11
	Closed       LifecycleState = 12
)
