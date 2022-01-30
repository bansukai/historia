package historia

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_AggregateRoot_TrackChange_should_add_and_transition_event(t *testing.T) {
	now := time.Now()
	id := "cupcakes"

	SetIDFunc(func() string { return id })
	SetNowFunc(func() time.Time { return now })

	defer func() {
		SetIDFunc(defaultIDGenerator)
		SetNowFunc(time.Now)
	}()

	name := "Someone Special"
	p := aggAgg{}
	p.TrackChange(&p, &born{Name: name})
	assert.Equal(t, id, p.ID())
	assert.Equal(t, p.name, name)
	assert.Len(t, p.events, 1)
	assert.Equal(t, Version(1), p.Version())

	e := p.events[0]
	assert.Equal(t, p.ID(), e.AggregateID)
	assert.Equal(t, Version(1), e.Version)
	assert.Equal(t, "aggAgg", e.AggregateType)
	assert.Equal(t, now, e.Timestamp)
}

func Test_AggregateRoot_BuildFromHistory_should_transition_and_set_id_version(t *testing.T) {
	p := aggAgg{}
	p.TrackChange(&p, &born{Name: "Something"})
	p.TrackChange(&p, &born{Name: "Poop"})
	p.TrackChange(&p, &agedOneYear{})
	assert.True(t, p.HasUnsavedEvents())

	p1 := aggAgg{}
	p1.BuildFromHistory(&p1, p.Events())

	assert.Equal(t, "Poop", p1.name)
	assert.Equal(t, 1, p1.age)
	assert.Equal(t, Version(3), p1.Version())
	assert.False(t, p1.HasUnsavedEvents())
}

func Test_AggregateRoot_SetID_should_return_error_when_id_is_already_set(t *testing.T) {
	p := aggAgg{
		AggregateRoot: AggregateRoot{id: "happy"},
	}

	assert.ErrorIs(t, p.SetID("nope"), ErrAggregateAlreadyExists)
}

func Test_AggregateRoot_SetID_should_change_the_id(t *testing.T) {
	p := aggAgg{}
	assert.NoError(t, p.SetID("yes"))
	assert.Equal(t, "yes", p.ID())
}

func Test_AggregateRoot_Root_should_return_pointer_to_self(t *testing.T) {
	p := aggAgg{}
	assert.Equal(t, &p.AggregateRoot, p.Root())
}

func Test_AggregateRoot_setInternals_should_set_id_version_and_clear_events(t *testing.T) {
	p := aggAgg{}
	p.TrackChange(&p, &born{})
	assert.True(t, p.HasUnsavedEvents())

	p.setInternals("hello", 2)
	assert.Equal(t, "hello", p.ID())
	assert.Equal(t, Version(2), p.Version())
}

func Test_AggregateRoot_update_should_set_version_to_last_event_and_clear_events(t *testing.T) {
	p := aggAgg{}
	p.version = 20
	p.update()
	assert.Equal(t, Version(20), p.Version())

	p.events = append(p.events, Event{Version: 5}, Event{Version: 6})
	p.update()

	assert.Equal(t, Version(6), p.Version())
	assert.Empty(t, p.events)
}

func Test_AggregateRoot_path_should_return_path(t *testing.T) {
	p := aggAgg{}
	assert.Equal(t, "github.com/reiroldan/historia", p.path())
}

// region mocks

type aggAgg struct {
	AggregateRoot

	name string
	age  int
}

type born struct{ Name string }
type agedOneYear struct{}

func (p *aggAgg) Transition(evt Event) {
	switch e := evt.Data.(type) {
	case *born:
		p.age = 0
		p.name = e.Name
	case *agedOneYear:
		p.age++
	}
}

func (p *aggAgg) GrowOlder() {
	metaData := EventMetadata{
		"foo": "bar",
	}

	p.TrackChangeWithMetadata(p, &agedOneYear{}, metaData)
}

// endregion
