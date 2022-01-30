package historia

import (
	"errors"
	"reflect"
)

// ErrAggregateAlreadyExists returned if the aggregateID is set more than one time
var ErrAggregateAlreadyExists = errors.New("its not possible to set ID on already existing aggregate")

// Version represents a version number
type Version uint64

const (
	emptyAggregateID = ""
)

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	id      string
	version Version
	events  []Event
}

// ID returns the aggregate ID as a string
func (a *AggregateRoot) ID() string {
	return a.id
}

// SetID opens up the possibility to set manual aggregate ID from the outside
func (a *AggregateRoot) SetID(id string) error {
	if a.id != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}
	a.id = id
	return nil
}

// Version return the version based on events that are not stored
func (a *AggregateRoot) Version() Version {
	if len(a.events) == 0 {
		return a.version
	}

	return a.events[len(a.events)-1].Version
}

// Events return the aggregate events from the aggregate
// make a copy of the slice preventing outsiders modifying events.
func (a *AggregateRoot) Events() []Event {
	e := make([]Event, len(a.events))
	copy(e, a.events)
	return e
}

// HasUnsavedEvents return if there are unsaved events
func (a *AggregateRoot) HasUnsavedEvents() bool {
	return len(a.events) > 0
}

// TrackChange is used internally by behaviour methods to apply state changes for later persistence.
func (a *AggregateRoot) TrackChange(aggregate Aggregate, data interface{}) {
	a.TrackChangeWithMetadata(aggregate, data, nil)
}

// TrackChangeWithMetadata is used internally by behaviour methods to apply state changes for later persistence.
// metadata is handled by this func to store unrelated application state.
func (a *AggregateRoot) TrackChangeWithMetadata(aggregate Aggregate, data interface{}, metadata map[string]interface{}) {
	// This can be overwritten in the constructor of the aggregate
	if a.id == emptyAggregateID {
		a.id = idFunc()
	}

	name := reflect.TypeOf(aggregate).Elem().Name()
	event := Event{
		AggregateID:   a.id,
		Version:       a.nextVersion(),
		AggregateType: name,
		Timestamp:     timeNow(),
		Data:          data,
		Metadata:      metadata,
	}
	a.events = append(a.events, event)
	aggregate.Transition(event)
}

// BuildFromHistory builds the aggregate state from events
func (a *AggregateRoot) BuildFromHistory(aggregate Aggregate, events []Event) {
	for _, event := range events {
		aggregate.Transition(event)
		a.id = event.AggregateID
		a.version = event.Version
	}
}

// Root returns the included Aggregate Root state, and is used from the interface Aggregate.
func (a *AggregateRoot) Root() *AggregateRoot {
	return a
}

func (a *AggregateRoot) setInternals(id string, version Version) {
	a.id = id
	a.version = version
	a.events = []Event{}
}

func (a *AggregateRoot) nextVersion() Version {
	return a.Version() + 1
}

// update sets the Version to the values in the last event.
// This function is called after the aggregate is saved in the repository
func (a *AggregateRoot) update() {
	if len(a.events) == 0 {
		return
	}

	lastEvent := a.events[len(a.events)-1]
	a.version = lastEvent.Version
	a.events = []Event{}
}

// path return the full name of the aggregate making it unique to other aggregates with
// the same name but placed in other packages.
func (a *AggregateRoot) path() string {
	return reflect.TypeOf(a).Elem().PkgPath()
}
