package historia

import (
	"github.com/google/uuid"
)

// idFunc is a global function that generates aggregate id's.
// It can be changed from the outside via the SetIDFunc function.
var idFunc = defaultIDGenerator

// NewID generates a unique id using the configured generator.
func NewID() string {
	return idFunc()
}

// SetIDFunc is used to change how aggregate IDs are generated.
func SetIDFunc(f func() string) {
	idFunc = f
}

func defaultIDGenerator() string {
	return uuid.NewString()
}
