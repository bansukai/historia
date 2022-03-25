package historia

import (
	"context"
	"errors"
	"time"
)

var (
	ErrAggregateMissingID              = errors.New("aggregate id is empty")
	ErrUnsavedEvents                   = errors.New("aggregate holds unsaved events")
	ErrAggregateDoesntSupportSnapshots = errors.New("aggregate doesn't implement SnapshotTaker interface")
)

type Snapshot struct {
	ID        string
	Timestamp time.Time
	Type      string
	State     []byte
	Version   Version
}

type SnapshotStore interface {
	Get(ctx context.Context, aggregateID string, aggregateType string) (*Snapshot, error)
	Save(ctx context.Context, ss *Snapshot) error
}

type SnapshotBody interface{}

type SnapshotTaker interface {
	TakeSnapshot() SnapshotBody
	ApplySnapshot(state SnapshotBody) error
}

type Marshaller interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// NewSnapper creates and returns an instance of Snapper
func NewSnapper(ss SnapshotStore, m Marshaller) *Snapper {
	return &Snapper{
		store:      ss,
		marshaller: m,
	}
}

// Snapper saves/applies snapshots to/from Aggregate
type Snapper struct {
	store      SnapshotStore
	marshaller Marshaller
}

func (s *Snapper) ApplySnapshot(ctx context.Context, aggregateID string, aggregate Aggregate) error {
	st, ok := aggregate.(SnapshotTaker)
	if !ok {
		return ErrAggregateDoesntSupportSnapshots
	}

	t := formatAggregatePathType(aggregate)
	snap, err := s.store.Get(ctx, aggregateID, t)
	if err != nil {
		return err
	}

	var state SnapshotBody
	if err := s.marshaller.Unmarshal(snap.State, &state); err != nil {
		return err
	}

	if err := st.ApplySnapshot(&state); err != nil {
		return err
	}

	root := aggregate.Root()
	root.setInternals(snap.ID, snap.Version)
	return nil
}

func (s *Snapper) SaveSnapshot(ctx context.Context, aggregate Aggregate) error {
	root := aggregate.Root()
	if err := validate(*root); err != nil {
		return err
	}

	st, ok := aggregate.(SnapshotTaker)
	if !ok {
		return ErrAggregateDoesntSupportSnapshots
	}

	payload := st.TakeSnapshot()
	if payload == nil {
		return nil
	}

	typ := formatAggregatePathType(aggregate)
	buf, err := s.marshaller.Marshal(payload)
	if err != nil {
		return err
	}

	snap := Snapshot{
		ID:        root.ID(),
		Timestamp: timeNow(),
		Type:      typ,
		Version:   root.Version(),
		State:     buf,
	}

	return s.store.Save(ctx, &snap)
}

// validate make sure the aggregate is valid to be saved
func validate(root AggregateBase) error {
	if root.ID() == "" {
		return ErrAggregateMissingID
	}

	if root.HasUnsavedEvents() {
		return ErrUnsavedEvents
	}

	return nil
}
