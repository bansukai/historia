package historia

import (
	"errors"
	"reflect"
	"time"
)

var (
	// ErrNoEvents when there is no events to get
	ErrNoEvents = errors.New("no events")

	// ErrNoMoreEvents when iterator has no more events to deliver
	ErrNoMoreEvents = errors.New("no more events")
)

type EventData interface{}

type EventMetadata map[string]interface{}

// Event holding metadata and the application specific event in the Data property
type Event struct {
	AggregateID   string
	AggregateType string
	Version       Version
	Timestamp     time.Time
	Data          EventData
	Metadata      EventMetadata
}

// Reason returns the name of the Data field
func (e Event) Reason() string {
	if e.Data == nil {
		return ""
	}
	return reflect.TypeOf(e.Data).Elem().Name()
}
