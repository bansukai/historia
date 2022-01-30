package historia

import (
	"errors"
	"reflect"
)

var (
	ErrAggregateMissingID = errors.New("aggregate id is empty")
	ErrUnsavedEvents      = errors.New("aggregate holds unsaved events")
)

type Snapshot struct {
	ID      string
	Type    string
	State   []byte
	Version Version
}

type SnapshotStore interface {
	Get(id string, t string) (Snapshot, error)
	Save(ss Snapshot) error
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
	typ := reflect.TypeOf(a).Elem().Name()
	snap, err := s.store.Get(aggregateID, typ)
	if err != nil {
		return err
	}

	if err = s.marshaller.Unmarshal(snap.State, a); err != nil {
		return err
	}

	root := a.Root()
	root.setInternals(snap.ID, snap.Version)
	return nil
}

// Save transform an aggregate to a snapshot
func (s *Snapper) Save(a Aggregate) error {
	root := a.Root()
	if err := validate(*root); err != nil {
		return err
	}

	typ := reflect.TypeOf(a).Elem().Name()
	b, err := s.marshaller.Marshal(a)
	if err != nil {
		return err
	}

	snap := Snapshot{
		ID:      root.ID(),
		Type:    typ,
		Version: root.Version(),
		State:   b,
	}

	return s.store.Save(snap)
}

// validate make sure the aggregate is valid to be saved
func validate(root AggregateRoot) error {
	if root.ID() == "" {
		return ErrAggregateMissingID
	}

	if root.HasUnsavedEvents() {
		return ErrUnsavedEvents
	}

	return nil
}
