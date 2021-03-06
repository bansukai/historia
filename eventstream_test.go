package historia

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEventStream_should_return_EventStream(t *testing.T) {
	es := NewEventStream()
	assert.NotNil(t, es.aggregateTypes)
	assert.NotNil(t, es.specificAggregates)
	assert.NotNil(t, es.specificEvents)
	assert.NotNil(t, es.allEvents)
}

func Test_EventStream_formatAggregatePathName(t *testing.T) {
	p := formatAggregatePathType(&esAgg{})
	assert.Equal(t, "github.com/bansukai/historia.esAgg", p)
}

func Test_EventStream_formatAggregatePathNameID(t *testing.T) {
	p := formatAggregatePathNameID(&esAgg{AggregateBase{id: "yay"}})
	assert.Equal(t, "github.com/bansukai/historia.esAgg#yay", p)
}

func Test_EventStream_SubscribeAll(t *testing.T) {
	var streamEvent *Event
	es := NewEventStream()
	s := es.SubscriberAll(func(ctx context.Context, e Event) error {
		streamEvent = &e
		return nil
	})

	s.Subscribe()
	assert.Len(t, es.allEvents, 1)

	var event = Event{Version: 123, Data: &esEvent{Name: "123"}, AggregateType: "esAgg"}
	assert.NoError(t, es.Update(context.Background(), &esAgg{}, []Event{event}))

	assert.NotNil(t, streamEvent)
	assert.Equal(t, event.Version, streamEvent.Version)

	s.Unsubscribe()
	assert.Len(t, es.allEvents, 0)
}

func Test_EventStream_SubscribeSpecificEvent(t *testing.T) {
	var streamEvent *Event
	es := NewEventStream()
	s := es.SubscriberSpecificEvent(func(ctx context.Context, e Event) error {
		streamEvent = &e
		return nil
	}, &esEvent{})
	s.Subscribe()
	assert.Len(t, es.specificEvents[reflect.TypeOf(&esEvent{})], 1)

	var event = Event{Version: 123, Data: &esEvent{Name: "123"}, AggregateType: "esAgg"}
	assert.NoError(t, es.Update(context.Background(), &esAgg{}, []Event{event}))

	assert.NotNil(t, streamEvent)
	assert.Equal(t, event.Version, streamEvent.Version)

	s.Unsubscribe()
	assert.Len(t, es.specificEvents[reflect.TypeOf(&esEvent{})], 0)
}

func Test_EventStream_SubscribeSpecificAggregateType(t *testing.T) {
	var streamEvent *Event
	es := NewEventStream()
	s := es.SubscriberAggregateType(func(ctx context.Context, e Event) error {
		streamEvent = &e
		return nil
	}, &esAgg{}, &esAggOther{})

	s.Subscribe()
	assert.Len(t, es.aggregateTypes, 2)
	assert.Len(t, es.aggregateTypes[formatAggregatePathType(&esAgg{})], 1)
	assert.Len(t, es.aggregateTypes[formatAggregatePathType(&esAggOther{})], 1)

	// update with event from the AnAggregate aggregate
	var event1 = Event{Version: 123, Data: &esEvent{Name: "123"}, AggregateType: "esAgg"}
	assert.NoError(t, es.Update(context.Background(), &esAgg{}, []Event{event1}))
	assert.NotNil(t, streamEvent)
	assert.Equal(t, event1.Version, streamEvent.Version)

	// update with event from the AnotherAggregate aggregate
	var event2 = Event{Version: 123, Data: &esEvent{Name: "Moo"}, AggregateType: "esAggOther"}
	assert.NoError(t, es.Update(context.Background(), &esAggOther{}, []Event{event2}))
	assert.Equal(t, event2.Version, streamEvent.Version)

	s.Unsubscribe()
	assert.Len(t, es.aggregateTypes[formatAggregatePathType(&esAgg{})], 0)
	assert.Len(t, es.aggregateTypes[formatAggregatePathType(&esAggOther{})], 0)
}

func Test_EventStream_SubscriberSpecificAggregate(t *testing.T) {
	first := esAgg{AggregateBase: AggregateBase{id: "123"}}
	second := esAggOther{AggregateBase: AggregateBase{id: "abc"}}

	var streamEvent *Event
	es := NewEventStream()
	s := es.SubscriberSpecificAggregate(func(ctx context.Context, e Event) error {
		streamEvent = &e
		return nil
	}, &first, &second)
	s.Subscribe()
	assert.Len(t, es.specificAggregates, 2)
	assert.Len(t, es.specificAggregates[formatAggregatePathNameID(&first)], 1)
	assert.Len(t, es.specificAggregates[formatAggregatePathNameID(&second)], 1)

	// update with event1 from the esAgg aggregate
	var event1 = Event{Version: 123, Data: &esEvent{Name: "Poo"}, AggregateType: "esAgg"}
	assert.NoError(t, es.Update(context.Background(), &first, []Event{event1}))
	assert.NotNil(t, streamEvent)
	assert.Equal(t, event1.Version, streamEvent.Version)

	// update with event2 from the esAggOther aggregate
	var event2 = Event{Version: 123, Data: &esEvent{Name: "Moo"}, AggregateType: "esAggOther"}
	assert.NoError(t, es.Update(context.Background(), &second, []Event{event2}))
	assert.Equal(t, event2.Version, streamEvent.Version)

	s.Unsubscribe()
	assert.Len(t, es.specificAggregates[formatAggregatePathNameID(&first)], 0)
	assert.Len(t, es.specificAggregates[formatAggregatePathNameID(&second)], 0)
}

func Test_EventStream_Multiple(t *testing.T) {
	streamEvent1 := make([]Event, 0)
	streamEvent2 := make([]Event, 0)
	streamEvent3 := make([]Event, 0)
	streamEvent4 := make([]Event, 0)
	streamEvent5 := make([]Event, 0)

	es := NewEventStream()

	type anEvent struct{ Name string }
	type anotherEvent struct{}

	es.SubscriberSpecificEvent(func(ctx context.Context, e Event) error {
		streamEvent1 = append(streamEvent1, e)
		return nil
	}, &anotherEvent{}).Subscribe()

	es.SubscriberSpecificEvent(func(ctx context.Context, e Event) error {
		streamEvent2 = append(streamEvent2, e)
		return nil
	}, &anotherEvent{}, &anEvent{}).Subscribe()

	es.SubscriberSpecificEvent(func(ctx context.Context, e Event) error {
		streamEvent3 = append(streamEvent3, e)
		return nil
	}, &anEvent{}).Subscribe()

	es.SubscriberAll(func(ctx context.Context, e Event) error {
		streamEvent4 = append(streamEvent4, e)
		return nil
	}).Subscribe()

	es.SubscriberAggregateType(func(ctx context.Context, e Event) error {
		streamEvent5 = append(streamEvent5, e)
		return nil
	}, &esAgg{}).Subscribe()

	var event = Event{Version: 123, Data: &anEvent{Name: "Poo"}, AggregateType: "esAgg"}
	assert.NoError(t, es.Update(context.Background(), &esAgg{}, []Event{event}))

	assert.Len(t, streamEvent1, 0, "stream1 should not have any events")
	assert.Len(t, streamEvent2, 1, "stream2 should have one event")
	assert.Len(t, streamEvent3, 1, "stream3 should have one event")
	assert.Len(t, streamEvent4, 1, "stream4 should have one event")
	assert.Len(t, streamEvent5, 1, "stream5 should have one event")
}

// region Mocks

type esEvent struct{ Name string }

type esAgg struct{ AggregateBase }

func (e *esAgg) Transition(_ Event) {}

type esAggOther struct{ AggregateBase }

func (e *esAggOther) Transition(_ Event) {}

// endregion
