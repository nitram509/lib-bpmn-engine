package BPMN20

// ActivityState as per BPMN 2.0 spec, section 13.2.2 Activity, page 428, State diagram:
//
//	              (Inactive)
//	                  O
//	                  |
//	A Token           v
//	Arrives        ┌─────┐
//	               │Ready│
//	               └─────┘
//	                  v         Activity Interrupted             An Alternative Path For
//	                  O -------------------------------------->O----------------------------+
//	Data InputSet     v                                        | Event Gateway Selected     |
//	Available     ┌──────┐                         Interrupting|                            |
//	              │Active│                         Event       |                            |
//	              └──────┘                                     |                            v
//	                  v         Activity Interrupted           v An Alternative Path For┌─────────┐
//	                  O -------------------------------------->O ---------------------->│Withdrawn│
//	Activity's work   v                                        | Event Gateway Selected └─────────┘
//	completed     ┌──────────┐                     Interrupting|                            |
//	              │Completing│                     Event       |                 The Process|
//	              └──────────┘                                 |                 Ends       |
//	                  v         Activity Interrupted           v  Non-Error                 |
//	Completing        O -------------------------------------->O--------------+             |
//	Requirements Done v                                  Error v              v             |
//	Assignments   ┌─────────┐                              ┌───────┐       ┌───────────┐    |
//	Completed     │Completed│                              │Failing│       │Terminating│    |
//	              └─────────┘                              └───────┘       └───────────┘    |
//	                  v  Compensation ┌────────────┐          v               v             |
//	                  O ------------->│Compensating│          O <-------------O Terminating |
//	                  |  Occurs       └────────────┘          v               v Requirements Done
//	      The Process |         Compensation v   Compensation  |           ┌──────────┐     |
//	      Ends        |       +--------------O----------------/|\--------->│Terminated│     |
//	                  |       | Completes    |   Interrupted   |           └──────────┘     |
//	                  |       v              |                 v              |             |
//	                  | ┌───────────┐        |Compensation┌──────┐            |             |
//	                  | │Compensated│        +----------->│Failed│            |             |
//	                  | └─────┬─────┘         Failed      └──────┘            |             |
//	                  |       |                               |               |             |
//	                  v      / The Process Ends               / Process Ends /              |
//	                  O<--------------------------------------------------------------------+
//	             (Closed)
type ActivityState string

const (
	Active       ActivityState = "ACTIVE"
	Compensated  ActivityState = "COMPENSATED"
	Compensating ActivityState = "COMPENSATING"
	Completed    ActivityState = "COMPLETED"
	Completing   ActivityState = "COMPLETING"
	Failed       ActivityState = "FAILED"
	Failing      ActivityState = "FAILING"
	Ready        ActivityState = "READY"
	Terminated   ActivityState = "TERMINATED"
	Terminating  ActivityState = "TERMINATING"
	Withdrawn    ActivityState = "WITHDRAWN"
)
