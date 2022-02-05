package historia

import (
	"errors"
	"time"
)

var (
	// ErrNoEvents when there are no events to apply
	ErrNoEvents = errors.New("no events")
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
	return TypeOf(e.Data)
}
