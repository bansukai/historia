package historia

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_NewSnapper_should_return_valid_instance(t *testing.T) {
	sm := &snapStoreMocker{}
	m := &marshalMocker{}
	s := NewSnapper(sm, m)

	assert.Equal(t, sm, s.store)
	assert.Equal(t, m, s.marshaller)
}

func Test_Snapper_Get_should_return_error_when_aggregate_doesnt_implement_SnapshotTaker(t *testing.T) {
	s := NewSnapper(nil, nil)
	assert.ErrorIs(t, s.Get("d", &ssAggNoSnapshot{}), ErrAggregateDoesntSupportSnapshots)
}

func Test_Snapper_Get_should_return_snapStore_Get_error(t *testing.T) {
	err := errors.New("nope")
	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return Snapshot{}, err },
	}
	s := NewSnapper(sm, nil)

	assert.ErrorIs(t, s.Get("poop", &ssAggWithSnapshot{}), err)
}

func Test_Snapper_Get_should_return_unmarshal_error(t *testing.T) {
	err := errors.New("nope")
	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return Snapshot{}, nil },
	}
	m := &marshalMocker{
		unmarshal: func(data []byte, v interface{}) error { return err },
	}

	s := NewSnapper(sm, m)

	assert.ErrorIs(t, s.Get("poop", &ssAggWithSnapshot{}), err)
}

func Test_Snapper_Get_should_return_SnapshotTaker_ApplySnapshot_error(t *testing.T) {
	err := errors.New("nope")
	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return Snapshot{}, nil },
	}
	m := &marshalMocker{
		unmarshal: func(data []byte, v interface{}) error { return nil },
	}

	s := NewSnapper(sm, m)
	agg := &ssAggWithSnapshot{
		applySnapshot: func(state SnapshotBody) error { return err },
	}

	assert.ErrorIs(t, s.Get("poop", agg), err)
}

func Test_Snapper_Get_should_call_setInternals(t *testing.T) {
	snap := Snapshot{
		ID:      "some_id",
		Type:    "",
		State:   nil,
		Version: 20,
	}

	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return snap, nil },
	}

	m := &marshalMocker{
		unmarshal: func(data []byte, v interface{}) error { return nil },
	}

	agg := &ssAggWithSnapshot{
		applySnapshot: func(SnapshotBody) error { return nil },
	}

	s := NewSnapper(sm, m)
	assert.NoError(t, s.Get(snap.ID, agg))

	assert.Equal(t, snap.ID, agg.ID())
	assert.Equal(t, snap.Version, agg.Version())
}

func Test_Snapper_Save_should_return_error_when_aggregate_doesnt_have_id(t *testing.T) {
	s := NewSnapper(nil, nil)
	assert.ErrorIs(t, s.Save(&ssAggWithSnapshot{}), ErrAggregateMissingID)
}

func Test_Snapper_Save_should_return_error_when_aggregate_has_unsaved_events(t *testing.T) {
	s := NewSnapper(nil, nil)
	agg := &ssAggWithSnapshot{
		AggregateRoot: AggregateRoot{
			id:     "yes",
			events: []Event{{}},
		},
	}
	assert.ErrorIs(t, s.Save(agg), ErrUnsavedEvents)
}

func Test_Snapper_Save_should_return_error_when_aggregate_doesnt_implement_SnapshotTaker(t *testing.T) {
	err := errors.New("poop")
	m := &marshalMocker{
		marshal: func(v interface{}) ([]byte, error) { return nil, err },
	}

	s := NewSnapper(nil, m)
	agg := &ssAggNoSnapshot{
		AggregateRoot: AggregateRoot{
			id: "yes",
		},
	}

	assert.ErrorIs(t, s.Save(agg), ErrAggregateDoesntSupportSnapshots)
}

func Test_Snapper_Save_should_return_if_TakeSnapshot_returns_nil(t *testing.T) {
	agg := &ssAggWithSnapshot{
		AggregateRoot: AggregateRoot{
			id:      "yes",
			version: Version(12),
		},
		takeSnapshot: func() SnapshotBody { return nil },
	}

	s := NewSnapper(nil, nil)
	assert.NoError(t, s.Save(agg))
}

func Test_Snapper_Save_should_return_error_when_marshal_returns_error(t *testing.T) {
	agg := &ssAggWithSnapshot{
		AggregateRoot: AggregateRoot{
			id:      "yes",
			version: Version(12),
		},
		takeSnapshot: func() SnapshotBody { return []byte{} },
	}

	err := errors.New("nope")
	m := &marshalMocker{
		marshal: func(v interface{}) ([]byte, error) { return nil, err },
	}

	s := NewSnapper(nil, m)
	assert.ErrorIs(t, s.Save(agg), err)
}

func Test_Snapper_Save_should_call_save_on_snapStore(t *testing.T) {
	now := time.Now()
	SetNowFunc(func() time.Time { return now })
	defer func() { SetNowFunc(time.Now) }()

	err := errors.New("poop")
	agg := &ssAggWithSnapshot{
		AggregateRoot: AggregateRoot{
			id:      "yes",
			version: Version(12),
		},
		takeSnapshot: func() SnapshotBody { return struct{}{} },
	}

	state := []byte{1, 2, 3}

	m := &marshalMocker{
		marshal: func(v interface{}) ([]byte, error) { return state, nil },
	}

	ss := &snapStoreMocker{
		save: func(ss Snapshot) error {
			assert.Equal(t, agg.ID(), ss.ID)
			assert.Equal(t, now, ss.Timestamp)
			assert.Equal(t, "ssAggWithSnapshot", ss.Type)
			assert.Equal(t, state, ss.State)
			assert.Equal(t, agg.Version(), ss.Version)
			return err
		},
	}

	s := NewSnapper(ss, m)

	assert.ErrorIs(t, s.Save(agg), err)
}

// region mocks

type ssAggNoSnapshot struct {
	AggregateRoot
}

func (s *ssAggNoSnapshot) Transition(Event) {}

type ssAggWithSnapshot struct {
	AggregateRoot
	takeSnapshot  func() SnapshotBody
	applySnapshot func(state SnapshotBody) error
}

func (s *ssAggWithSnapshot) TakeSnapshot() SnapshotBody             { return s.takeSnapshot() }
func (s *ssAggWithSnapshot) ApplySnapshot(state SnapshotBody) error { return s.applySnapshot(state) }
func (s *ssAggWithSnapshot) Transition(Event)                       {}

type snapStoreMocker struct {
	get  func(aggregateID, t string) (Snapshot, error)
	save func(ss Snapshot) error
}

func (s *snapStoreMocker) Get(aggregateID, t string) (Snapshot, error) { return s.get(aggregateID, t) }
func (s *snapStoreMocker) Save(ss Snapshot) error                      { return s.save(ss) }

type marshalMocker struct {
	marshal   func(v interface{}) ([]byte, error)
	unmarshal func(data []byte, v interface{}) error
}

func (m *marshalMocker) Marshal(v interface{}) ([]byte, error)      { return m.marshal(v) }
func (m *marshalMocker) Unmarshal(data []byte, v interface{}) error { return m.unmarshal(data, v) }

// endregion
