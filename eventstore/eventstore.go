package eventstore

import (
	"errors"

	hi "github.com/bansukai/historia"
)

var (
	// ErrEventMultipleAggregates when events holds different id
	ErrEventMultipleAggregates = errors.New("events holds events for more than one aggregate")

	// ErrEventMultipleAggregateTypes when events hold different aggregate types
	ErrEventMultipleAggregateTypes = errors.New("events holds events for more than one aggregate type")

	// ErrConcurrency when the currently saved version of the aggregate differs from the new ones
	ErrConcurrency = errors.New("concurrency error")

	// ErrReasonMissing when the reason is not present in the events
	ErrReasonMissing = errors.New("event holds no reason")
)

// ValidateEvents make sure the incoming events are valid
func ValidateEvents(aggregateID string, currentVersion hi.Version, events []hi.Event) error {
	at := events[0].AggregateType

	for i := range events {
		event := events[i]
		if event.AggregateID != aggregateID {
			return ErrEventMultipleAggregates
		}

		if event.AggregateType != at {
			return ErrEventMultipleAggregateTypes
		}

		if currentVersion+1 != event.Version {
			return ErrConcurrency
		}

		if event.Reason() == "" {
			return ErrReasonMissing
		}

		currentVersion = event.Version
	}
	return nil
}
