package historia

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// NewEventStream factory function
func NewEventStream() *EventStream {
	return &EventStream{
		aggregateTypes:     make(map[string][]*Subscription),
		specificAggregates: make(map[string][]*Subscription),
		specificEvents:     make(map[reflect.Type][]*Subscription),
		allEvents:          []*Subscription{},
	}
}

// Subscription holding the subscribe / unsubscribe / and func to be called when event matches the subscription
type Subscription struct {
	f EventHandlerFunc
	u func()
	s func()
}

// Unsubscribe invokes the unsubscribe function
func (s *Subscription) Unsubscribe() {
	s.u()
}

// Subscribe invokes the subscribe function
func (s *Subscription) Subscribe() {
	s.s()
}

// EventStream holds event subscriptions
type EventStream struct {
	aggregateTypes     map[string][]*Subscription
	specificAggregates map[string][]*Subscription
	specificEvents     map[reflect.Type][]*Subscription
	allEvents          []*Subscription

	lock sync.Mutex
}

// Update invoke all event handling functions for subscriptions
func (e *EventStream) Update(ctx context.Context, aggregate Aggregate, events []Event) error {
	e.lock.Lock()
	defer e.lock.Unlock()

	for i := range events {
		event := events[i]

		// call all functions that has registered for all events
		if err := updateAll(ctx, e, event); err != nil {
			return err
		}

		// call all functions that has registered for the specific event
		if err := updateSpecificEvent(ctx, e, event); err != nil {
			return err
		}

		// call all functions that has registered for the aggregate type events
		if err := updateSpecificAggregateEvents(ctx, aggregate, e, event); err != nil {
			return err
		}

		// call all functions that has registered for the aggregate type and ID events
		if err := updateSpecificAggregate(ctx, aggregate, e, event); err != nil {
			return err
		}
	}
	return nil
}

// SubscriberAll bind a function to be called on all events
func (e *EventStream) SubscriberAll(f EventHandlerFunc) *Subscription {
	s := Subscription{
		f: f,
	}
	s.u = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i, sub := range e.allEvents {
			if &s == sub {
				e.allEvents = append(e.allEvents[:i], e.allEvents[i+1:]...)
				break
			}
		}
	}
	s.s = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.allEvents = append(e.allEvents, &s)
	}
	return &s
}

// SubscriberSpecificAggregate bind a function to be called on events that happen on an aggregate based on type and ID
func (e *EventStream) SubscriberSpecificAggregate(f EventHandlerFunc, aggregates ...Aggregate) *Subscription {
	s := Subscription{
		f: f,
	}
	s.u = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for x := range aggregates {
			a := aggregates[x]
			ref := formatAggregatePathNameID(a)
			for i, sub := range e.specificAggregates[ref] {
				if &s == sub {
					e.specificAggregates[ref] = append(e.specificAggregates[ref][:i], e.specificAggregates[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.s = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i := range aggregates {
			a := aggregates[i]
			ref := formatAggregatePathNameID(a)
			e.specificAggregates[ref] = append(e.specificAggregates[ref], &s)
		}
	}
	return &s
}

// SubscriberAggregateType bind a function to be called on events for an aggregate type
func (e *EventStream) SubscriberAggregateType(f EventHandlerFunc, aggregates ...Aggregate) *Subscription {
	s := Subscription{
		f: f,
	}
	s.u = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for x := range aggregates {
			a := aggregates[x]
			ref := formatAggregatePathType(a)
			for i, sub := range e.aggregateTypes[ref] {
				if &s == sub {
					e.aggregateTypes[ref] = append(e.aggregateTypes[ref][:i], e.aggregateTypes[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.s = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i := range aggregates {
			ref := formatAggregatePathType(aggregates[i])
			e.aggregateTypes[ref] = append(e.aggregateTypes[ref], &s)
		}
	}
	return &s
}

// SubscriberSpecificEvent bind a function to be called on specific events
func (e *EventStream) SubscriberSpecificEvent(f EventHandlerFunc, events ...EventData) *Subscription {
	s := Subscription{
		f: f,
	}
	s.u = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for x := range events {
			event := events[x]
			t := reflect.TypeOf(event)
			for i, sub := range e.specificEvents[t] {
				if &s == sub {
					e.specificEvents[t] = append(e.specificEvents[t][:i], e.specificEvents[t][i+1:]...)
					break
				}
			}
		}
	}
	s.s = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i := range events {
			t := reflect.TypeOf(events[i])
			e.specificEvents[t] = append(e.specificEvents[t], &s)
		}
	}
	return &s
}

func updateAll(ctx context.Context, stream *EventStream, event Event) error {
	for i := range stream.allEvents {
		if err := stream.allEvents[i].f(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func updateSpecificEvent(ctx context.Context, stream *EventStream, event Event) error {
	t := reflect.TypeOf(event.Data)
	if subs, ok := stream.specificEvents[t]; ok {
		for i := range subs {
			if err := subs[i].f(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateSpecificAggregateEvents(ctx context.Context, aggregate Aggregate, stream *EventStream, event Event) error {
	ref := formatAggregatePathType(aggregate)
	if subs, ok := stream.aggregateTypes[ref]; ok {
		for i := range subs {
			if err := subs[i].f(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateSpecificAggregate(ctx context.Context, aggregate Aggregate, stream *EventStream, event Event) error {
	ref := formatAggregatePathNameID(aggregate)
	if subs, ok := stream.specificAggregates[ref]; ok {
		for i := range subs {
			if err := subs[i].f(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatAggregatePathType(aggregate Aggregate) string {
	root := PathOf(aggregate)
	name := TypeOf(aggregate)
	ref := fmt.Sprintf("%s.%s", root, name)
	return ref
}

func formatAggregatePathNameID(aggregate Aggregate) string {
	root := formatAggregatePathType(aggregate)
	ref := fmt.Sprintf("%s#%s", root, aggregate.Root().ID())
	return ref
}
