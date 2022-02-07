package historia

import (
	"context"
	"errors"
)

var (
	ErrNoSnapShotInitialized = errors.New("no snapshot store has been initialized")
	ErrSnapshotNotFound      = errors.New("snapshot not found")
	ErrAggregateNotFound     = errors.New("aggregate not found")
)

type EventHandlerFunc func(ctx context.Context, event Event) error

type Repository interface {
	// Get fetches the aggregates event and builds up the aggregate
	// If there are snapshots for the aggregate, it will attempt to apply them,
	// as well as all events after the version of the aggregate, if any.
	Get(ctx context.Context, aggregateID string, aggregate Aggregate) error

	// Save an aggregate's events
	Save(ctx context.Context, aggregate Aggregate) error

	// SubscriberAll bind a function to be called on all events
	SubscriberAll(f EventHandlerFunc) *Subscription

	// SubscriberSpecificAggregate bind a function to be called on events that happen on an aggregate based on type and ID
	SubscriberSpecificAggregate(f EventHandlerFunc, aggregates ...Aggregate) *Subscription

	// SubscriberAggregateType bind a function to be called on events for an aggregate type
	SubscriberAggregateType(f EventHandlerFunc, aggregates ...Aggregate) *Subscription

	// SubscriberSpecificEvent bind a function to be called on specific events
	SubscriberSpecificEvent(f EventHandlerFunc, events ...EventData) *Subscription
}

// EventStore interface exposes the methods an event store must uphold
type EventStore interface {

	// SaveEvents an aggregates events
	SaveEvents(ctx context.Context, events []Event) error

	// GetEvents events belonging to an aggregate
	GetEvents(ctx context.Context, aggregateID string, aggregateType string, afterVersion Version) ([]Event, error)

	// Close perform cleanup if required
	Close() error
}

// Aggregate interface to use the aggregate root specific methods
type Aggregate interface {
	Root() *AggregateBase
	Transition(evt Event)
}

type SnapShooter interface {
	// ApplySnapshot retrieves and applies snapshots onto the given Aggregate.
	ApplySnapshot(ctx context.Context, aggregateID string, aggregate Aggregate) error

	// SaveSnapshot requests a snapshot from the Aggregate,
	// which must implement the SnapshotTaker interface, and persists
	// it to the underlying SnapshotStore.
	SaveSnapshot(ctx context.Context, aggregate Aggregate) error
}

// NewRepository creates and returns a new instance of Repo
func NewRepository(es EventStore, s SnapShooter) *Repo {
	return &Repo{
		EventStream: NewEventStream(),
		eventStore:  es,
		snapper:     s,
	}
}

// Repo is the returned instance from the factory function
type Repo struct {
	*EventStream
	eventStore EventStore
	snapper    SnapShooter
}

// Get fetches the aggregates event and builds up the aggregate
// If there is a snapshot store, try to fetch a snapshot of the aggregate and
// event after the version of the aggregate, if any.
func (r *Repo) Get(ctx context.Context, aggregateID string, aggregate Aggregate) error {
	// if there is a snapshot store try fetch aggregate snapshot
	if r.snapper != nil {
		err := r.snapper.ApplySnapshot(ctx, aggregateID, aggregate)
		if err != nil && !errors.Is(err, ErrSnapshotNotFound) {
			return err
		}
	}

	// fetch events after the current version of the aggregate that could be fetched from the snapshot store
	root := aggregate.Root()
	aggregateType := TypeOf(aggregate)
	events, err := r.eventStore.GetEvents(ctx, aggregateID, aggregateType, root.Version())
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
func (r *Repo) Save(ctx context.Context, aggregate Aggregate) error {
	root := aggregate.Root()
	if err := r.eventStore.SaveEvents(ctx, root.events); err != nil {
		return err
	}

	// publish the saved events to subscribers
	if err := r.Update(ctx, aggregate, root.Events()); err != nil {
		return err
	}

	// update the internal aggregate state
	root.update()
	return nil
}

// SaveSnapshot saves the current state of the aggregate but only if it has no unsaved events
func (r *Repo) SaveSnapshot(ctx context.Context, aggregate Aggregate) error {
	if r.snapper == nil {
		return ErrNoSnapShotInitialized
	}

	return r.snapper.SaveSnapshot(ctx, aggregate)
}
