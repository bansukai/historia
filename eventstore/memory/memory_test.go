package memory

import (
	"testing"

	"github.com/reiroldan/historia/eventstore"
)

func TestMemoryStore(t *testing.T) {
	eventstore.AcceptanceTest(t, New())
}
