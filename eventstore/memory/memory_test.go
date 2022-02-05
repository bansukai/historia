package memory

import (
	"testing"

	"github.com/bansukai/historia/eventstore"
)

func TestMemoryStore(t *testing.T) {
	eventstore.AcceptanceTest(t, New())
}
