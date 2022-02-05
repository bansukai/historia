package memory

import (
	"context"
	"fmt"

	"github.com/bansukai/historia"
)

// New handler for the snapshot service
func New() *Memory {
	return &Memory{
		store: make(map[string]*historia.Snapshot),
	}
}

// Memory of snapshot store
type Memory struct {
	store map[string]*historia.Snapshot
}

func (h *Memory) Get(_ context.Context, aggregateID string, aggregateType string) (*historia.Snapshot, error) {
	k := formatSnapshotKey(aggregateID, aggregateType)
	v, ok := h.store[k]
	if !ok {
		return nil, historia.ErrSnapshotNotFound
	}
	return v, nil
}

func (h *Memory) Save(_ context.Context, ss *historia.Snapshot) error {
	k := formatSnapshotKey(ss.ID, ss.Type)
	h.store[k] = ss
	return nil
}

func formatSnapshotKey(id string, t string) string {
	return fmt.Sprintf("%s_%s", id, t)
}
