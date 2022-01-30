package historia

import (
	"errors"
	"reflect"
)

var (
	ErrNoSnapShotInitialized = errors.New("no snapshot store has been initialized")
	ErrSnapshotNotFound      = errors.New("snapshot not found")
	ErrAggregateNotFound     = errors.New("aggregate not found")
)

// EventStore interface exposes the methods an event store must uphold
type EventStore interface {
	Save(events []Event) error
	Get(id string, aggregateType string, afterVersion Version) ([]Event, error)
	Close() error
}

// Aggregate interface to use the aggregate root specific methods
type Aggregate interface {
	Root() *AggregateRoot
	Transition(evt Event)
}

type SnapShooter interface {
	Get(aggregateID string, a Aggregate) error
	Save(a Aggregate) error
}

func NewRepository(es EventStore, s SnapShooter) *Repository {
	return &Repository{
		eventStore:  es,
		snapper:     s,
		EventStream: NewEventStream(),
	}
}

// Repository is the returned instance from the factory function
type Repository struct {
	*EventStream
	eventStore EventStore
	snapper    SnapShooter
}

// Get fetches the aggregates event and builds up the aggregate
// If there is a snapshot store, try to fetch a snapshot of the aggregate and
// event after the version of the aggregate, if any.
func (r *Repository) Get(id string, aggregate Aggregate) error {
	// if there is a snapshot store try fetch aggregate snapshot
	if r.snapper != nil {
		err := r.snapper.Get(id, aggregate)
		if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
			return err
		}
	}

	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	root := aggregate.Root()
	aggregateType := reflect.TypeOf(aggregate).Elem().Name()
	events, err := r.eventStore.Get(id, aggregateType, root.Version())
	if err != nil {
		if !errors.Is(err, ErrNoEvents) {
			return err
		}
		if root.Version() == 0 {
			return ErrAggregateNotFound
		}
	}

	// apply the event on the aggregate
	root.BuildFromHistory(aggregate, events)
	return nil
}

// Save an aggregates events
func (r *Repository) Save(aggregate Aggregate) error {
	root := aggregate.Root()
	if err := r.eventStore.Save(root.events); err != nil {
		return err
	}

	// publish the saved events to subscribers
	r.Update(aggregate, root.Events())

	// update the internal aggregate state
	root.update()
	return nil
}

// SaveSnapshot saves the current state of the aggregate but only if it has no unsaved events
func (r *Repository) SaveSnapshot(aggregate Aggregate) error {
	if r.snapper == nil {
		return ErrNoSnapShotInitialized
	}

	return r.snapper.Save(aggregate)
}
