package memory

import (
	"testing"

	"github.com/reiroldan/historia/snapshot"
)

func TestStore(t *testing.T) {
	snapshot.AcceptanceTestSnapshotStore(t, New())
}
