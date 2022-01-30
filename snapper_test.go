package historia

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewSnapper_should_return_valid_instance(t *testing.T) {
	sm := &snapStoreMocker{}
	m := &marshalMocker{}
	s := NewSnapper(sm, m)

	assert.Equal(t, sm, s.store)
	assert.Equal(t, m, s.marshaller)
}

func Test_Snapper_Get_should_return_snapStore_error(t *testing.T) {
	err := errors.New("nope")
	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return Snapshot{}, err },
	}
	s := NewSnapper(sm, nil)

	assert.ErrorIs(t, s.Get("poop", &ssAgg{}), err)
}

func Test_Snapper_Get_should_return_marshal_error(t *testing.T) {
	err := errors.New("nope")
	sm := &snapStoreMocker{
		get: func(aggregateID, t string) (Snapshot, error) { return Snapshot{}, nil },
	}
	m := &marshalMocker{
		unmarshal: func(data []byte, v interface{}) error { return err },
	}

	s := NewSnapper(sm, m)

	assert.ErrorIs(t, s.Get("poop", &ssAgg{}), err)
}

func Test_Snapper_Get_should_call_setInternals(t *testing.T) {
	snap := Snapshot{
		ID:      "asdasdsad",
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

	agg := &ssAgg{}
	s := NewSnapper(sm, m)
	assert.NoError(t, s.Get(snap.ID, agg))

	assert.Equal(t, snap.ID, agg.ID())
	assert.Equal(t, snap.Version, agg.Version())
}

func Test_Snapper_Save_should_return_error_when_aggregate_doesnt_have_id(t *testing.T) {
	s := NewSnapper(nil, nil)
	assert.ErrorIs(t, s.Save(&ssAgg{}), ErrAggregateMissingID)
}

func Test_Snapper_Save_should_return_error_when_aggregate_has_unsaved_events(t *testing.T) {
	s := NewSnapper(nil, nil)
	agg := &ssAgg{
		AggregateRoot{
			id:     "yes",
			events: []Event{{}},
		},
	}
	assert.ErrorIs(t, s.Save(agg), ErrUnsavedEvents)
}

func Test_Snapper_Save_should_return_error_when_marshal_fails(t *testing.T) {
	err := errors.New("poop")
	m := &marshalMocker{
		marshal: func(v interface{}) ([]byte, error) { return nil, err },
	}

	s := NewSnapper(nil, m)
	agg := &ssAgg{
		AggregateRoot{
			id: "yes",
		},
	}

	assert.ErrorIs(t, s.Save(agg), err)
}

func Test_Snapper_Save_should_call_save_on_snapStore(t *testing.T) {
	err := errors.New("poop")
	agg := &ssAgg{
		AggregateRoot{
			id:      "yes",
			version: Version(12),
		},
	}

	state := []byte{1, 2, 3}

	m := &marshalMocker{
		marshal: func(v interface{}) ([]byte, error) { return state, nil },
	}

	ss := &snapStoreMocker{
		save: func(ss Snapshot) error {
			assert.Equal(t, agg.ID(), ss.ID)
			assert.Equal(t, "ssAgg", ss.Type)
			assert.Equal(t, state, ss.State)
			assert.Equal(t, agg.Version(), ss.Version)
			return err
		},
	}

	s := NewSnapper(ss, m)

	assert.ErrorIs(t, s.Save(agg), err)
}

// region mocks

type ssAgg struct {
	AggregateRoot
}

func (s *ssAgg) TakeSnapshot() (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ssAgg) ApplySnapshot(state interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (s *ssAgg) Transition(evt Event) {}

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
