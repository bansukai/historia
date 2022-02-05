package historia

import (
	"errors"
	"reflect"
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
	Get(id string, t string) (*Snapshot, error)
	Save(ss *Snapshot) error
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

func NewSnapper(ss SnapshotStore, m Marshaller) *Snapper {
	return &Snapper{
		store:      ss,
		marshaller: m,
	}
}

// Snapper gets and saves snapshots
type Snapper struct {
	store      SnapshotStore
	marshaller Marshaller
}

func (s *Snapper) Get(aggregateID string, a Aggregate) error {
	st, ok := a.(SnapshotTaker)
	if !ok {
		return ErrAggregateDoesntSupportSnapshots
	}

	t := reflect.TypeOf(a).Elem().Name()
	snap, err := s.store.Get(aggregateID, t)
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

	root := a.Root()
	root.setInternals(snap.ID, snap.Version)
	return nil
}

// Save requests a snapshot from the Aggregate, it must implement the SnapshotTaker interface.
func (s *Snapper) Save(a Aggregate) error {
	root := a.Root()
	if err := validate(*root); err != nil {
		return err
	}

	st, ok := a.(SnapshotTaker)
	if !ok {
		return ErrAggregateDoesntSupportSnapshots
	}

	payload := st.TakeSnapshot()
	if payload == nil {
		return nil
	}

	typ := reflect.TypeOf(a).Elem().Name()
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

	return s.store.Save(&snap)
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
