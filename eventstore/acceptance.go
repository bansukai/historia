package eventstore

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	hi "github.com/bansukai/historia"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func AcceptanceTest(t *testing.T, eventStore hi.EventStore) {

	tests := []struct {
		title string
		run   func(es hi.EventStore) error
	}{
		{"should save and get events", saveAndGetEvents},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			assert.NoError(t, test.run(eventStore))
		})
	}
}

//nolint:gocyclo // it's complicated
func saveAndGetEvents(es hi.EventStore) error {
	aggregateID := idFunc()
	events := createEvents(aggregateID)

	if err := es.SaveEvents(context.Background(), events); err != nil {
		return err
	}

	fetchedEvents, err := es.GetEvents(context.Background(), aggregateID, aggregateType, 0)
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
	err = es.SaveEvents(context.Background(), createEventsContinue(aggregateID))
	if err != nil {
		return err
	}

	fetchedEvents, err = es.GetEvents(context.Background(), aggregateID, aggregateType, 0)
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

var idFunc = uuid.NewString
var aggregateType = ""
var timestamp = time.Now()

type acceptanceAggregate struct {
	hi.AggregateBase
}

func (a *acceptanceAggregate) Transition(evt hi.Event) {}

type eventCreated struct {
	AccountID     string
	OpeningValue  int
	OpeningPoints int
}

type eventMatched struct {
	NewStatus int
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
		{AggregateID: aggregateID, Version: 1, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventCreated{AccountID: "1234567", OpeningValue: 10000, OpeningPoints: 0}, Metadata: metadata},
		{AggregateID: aggregateID, Version: 2, AggregateType: aggregateType, Timestamp: timestamp, Data: &eventMatched{NewStatus: 3}, Metadata: metadata},
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
