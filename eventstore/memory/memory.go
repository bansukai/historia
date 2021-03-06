package memory

import (
	"context"
	"sync"

	"github.com/bansukai/historia"
	"github.com/bansukai/historia/eventstore"
)

// New in memory event store
func New() *Memory {
	return &Memory{
		aggregateEvents: make(map[string][]historia.Event),
		allEvents:       make([]historia.Event, 0),
	}
}

// Memory is a handler for event streaming
type Memory struct {
	aggregateEvents map[string][]historia.Event
	allEvents       []historia.Event
	lock            sync.Mutex
}

// SaveEvents an aggregate (its events)
func (e *Memory) SaveEvents(ctx context.Context, events []historia.Event) error {
	// Return if there is no events to save
	if len(events) == 0 {
		return nil
	}

	// make sure its thread safe
	e.lock.Lock()
	defer e.lock.Unlock()

	// get bucket name from first event
	aggregateType := events[0].AggregateType
	aggregateID := events[0].AggregateID
	bucketName := aggregateKey(aggregateType, aggregateID)
	bucket := e.aggregateEvents[bucketName]
	currentVersion := historia.Version(0)

	if len(bucket) > 0 {
		currentVersion = bucket[len(bucket)-1].Version
	}

	err := eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return err
	}

	bucket = append(bucket, events...)
	e.aggregateEvents[bucketName] = bucket
	e.allEvents = append(e.allEvents, events...)
	return nil
}

// GetEvents aggregate events
func (e *Memory) GetEvents(ctx context.Context, aggregateID string, aggregateType string, afterVersion historia.Version) ([]historia.Event, error) {
	var events []historia.Event

	e.lock.Lock()
	defer e.lock.Unlock()

	aggEvents := e.aggregateEvents[aggregateKey(aggregateType, aggregateID)]
	for i := range aggEvents {
		event := aggEvents[i]
		if event.Version > afterVersion {
			events = append(events, event)
		}
	}

	if len(events) == 0 {
		return nil, historia.ErrNoEvents
	}

	return events, nil
}

// Close does nothing
func (e *Memory) Close() error {
	return nil
}

// aggregateKey generate an aggregate key to store events against from aggregateType and aggregateID
func aggregateKey(aggregateType, aggregateID string) string {
	return aggregateType + "_" + aggregateID
}
