package eventstore

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	hi "github.com/reiroldan/historia"
	"github.com/stretchr/testify/assert"
)

func AcceptanceTest(t *testing.T, eventStore hi.EventStore) {

	tests := []struct {
		title string
		run   func(es hi.EventStore) error
	}{
		{"should save and get events", saveAndGetEvents},
		//{"should get events after version", getEventsAfterVersion},
		//{"should not save events from different aggregates", saveEventsFromMoreThanOneAggregate},
		//{"should not save events from different aggregate types", saveEventsFromMoreThanOneAggregateType},
		//{"should not save events in wrong order", saveEventsInWrongOrder},
		//{"should not save events in wrong version", saveEventsInWrongVersion},
		//{"should not save event with no reason", saveEventsWithEmptyReason},
		//{"should save and get event concurrently", saveAndGetEventsConcurrently},
		//{"should return error when no events", getErrWhenNoEvents},
		//{"should get global event order from save", saveReturnGlobalEventOrder},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			assert.NoError(t, test.run(eventStore))
		})
	}
}

func saveAndGetEvents(es hi.EventStore) error {
	aggregateID := idFunc()
	events := createEvents(aggregateID)

	if err := es.Save(events); err != nil {
		return err
	}

	fetchedEvents, err := es.Get(aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}

	if len(fetchedEvents) != len(events) {
		return errors.New("wrong number of events returned")
	}

	if fetchedEvents[0].Version != events[0].Version {
		return errors.New("wrong events returned")
	}

	// Add more events
	events = append(events, createEventsContinue(aggregateID)...)

	// Add more events to the same aggregate event stream
	err = es.Save(createEventsContinue(aggregateID))
	if err != nil {
		return err
	}

	fetchedEvents, err = es.Get(aggregateID, aggregateType, 0)
	if err != nil {
		return err
	}

	if len(fetchedEvents) != len(events) {
		return errors.New("wrong number of events returned")
	}

	if fetchedEvents[0].Version != events[0].Version {
		return errors.New("wrong event version returned")
	}

	if fetchedEvents[len(fetchedEvents)-1].Version != events[len(events)-1].Version {
		return errors.New("wrong last event version returned")
	}

	if fetchedEvents[0].AggregateID != events[0].AggregateID {
		return errors.New("wrong event aggregateID returned")
	}

	if fetchedEvents[0].AggregateType != events[0].AggregateType {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEvents[0].Reason() != events[0].Reason() {
		return errors.New("wrong event aggregateType returned")
	}

	if fetchedEvents[0].Metadata["test"] != "hello" {
		return errors.New("wrong event meta data returned")
	}

	data, ok := fetchedEvents[0].Data.(*eventCreated)
	if !ok {
		return errors.New("wrong type in Data")
	}

	if data.OpeningValue != 10000 {
		return fmt.Errorf("wrong OpeningValue %d", data.OpeningValue)
	}

	return nil
}

var idFunc = func() string { return uuid.NewString() }
var aggregateType = ""
var timestamp = time.Now()

type status int

const (
	statusOne   = status(iota)
	statusTwo   = status(iota)
	statusThree = status(iota)
)

type acceptanceAggregate struct {
	hi.AggregateRoot
}

func (a *acceptanceAggregate) Transition(evt hi.Event) {}

type eventCreated struct {
	AccountId     string
	OpeningValue  int
	OpeningPoints int
}

type eventMatched struct {
	NewStatus status
}

type eventTaken struct {
	ValueAdded  int
	PointsAdded int
}

func createEvents(aggregateID string) []hi.Event {
	return createEventsWithID(aggregateID)
}

func createEventsWithID(aggregateID string) []hi.Event {
	metadata := make(map[string]interface{})
	metadata["test"] = "hello"
	history := []hi.Event{
		{AggregateID: aggregateID, Version: 1, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventCreated{AccountId: "1234567", OpeningValue: 10000, OpeningPoints: 0}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 2, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventMatched{NewStatus: statusTwo}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 3, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 2525, PointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 4, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 2512, PointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 5, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 5600, PointsAdded: 5}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 6, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 3000, PointsAdded: 3}, Metadata: metadata},
	}
	return history
}

func createEventsContinue(aggregateID string) []hi.Event {
	history := []hi.Event{
		{AggregateID: aggregateID, Version: 7, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 5600, PointsAdded: 5}},
		{AggregateID: aggregateID, Version: 8, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventTaken{ValueAdded: 3000, PointsAdded: 3}},
	}
	return history
}
