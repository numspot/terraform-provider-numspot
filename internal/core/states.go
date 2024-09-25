package core

const (
	attaching = "attaching"
	attached  = "attached"
	creating  = "creating"
	updating  = "updating"
	running   = "running"
	stopped   = "stopped"
	created   = "created"
	available = "available"
	inUse     = "in-use"
)

type (
	targetState  []string
	pendingState []string
)
