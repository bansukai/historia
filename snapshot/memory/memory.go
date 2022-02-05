package memory

import (
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

func (h *Memory) Get(id, t string) (*historia.Snapshot, error) {
	k := formatSnapshotKey(id, t)
	v, ok := h.store[k]
	if !ok {
		return nil, historia.ErrSnapshotNotFound
	}
	return v, nil
}

func (h *Memory) Save(s *historia.Snapshot) error {
	k := formatSnapshotKey(s.ID, s.Type)
	h.store[k] = s
	return nil
}

func formatSnapshotKey(id string, t string) string {
	return fmt.Sprintf("%s_%s", id, t)
}
