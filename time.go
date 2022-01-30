package historia

import (
	"time"
)

// timeNow is a global function that returns the current time.
// It can be changed from the outside via the SetNowFunc function.
var timeNow = time.Now

// SetNowFunc is used to change what time is returned.
func SetNowFunc(f func() time.Time) {
	timeNow = f
}
