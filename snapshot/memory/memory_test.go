package memory

import (
	"testing"

	"github.com/bansukai/historia/snapshot"
)

func TestStore(t *testing.T) {
	snapshot.AcceptanceTestSnapshotStore(t, New())
}
