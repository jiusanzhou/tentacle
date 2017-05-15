package mvc

import (
	"github.com/jiusanzhou/tentacle/util"
)

type Controller interface {
	// how the model communicates that it has changed state
	Update(State)

	// instructs the controller to shut the app down
	Shutdown(message string)

	// A channel of updates
	Updates() *util.Broadcast

	// returns the current state
	State() State

	// safe wrapper for running go-routines
	Go(fn func())
}
