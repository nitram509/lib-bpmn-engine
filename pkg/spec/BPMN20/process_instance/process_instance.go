package process_instance

type State string

const (
	READY     State = "READY"
	ACTIVE    State = "ACTIVE"
	COMPLETED State = "COMPLETED"
)
