package historia

import (
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
	f      func(e Event)
	unsubF func()
	subF   func()
}

// Unsubscribe invokes the unsubscribe function
func (s *Subscription) Unsubscribe() {
	s.unsubF()
}

// Subscribe invokes the subscribe function
func (s *Subscription) Subscribe() {
	s.subF()
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
func (e *EventStream) Update(aggregate Aggregate, events []Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, event := range events {
		// call all functions that has registered for all events
		for _, s := range e.allEvents {
			s.f(event)
		}

		// call all functions that has registered for the specific event
		t := reflect.TypeOf(event.Data)
		if subs, ok := e.specificEvents[t]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}

		ref := formatAggregatePathType(aggregate)
		// call all functions that has registered for the aggregate type events
		if subs, ok := e.aggregateTypes[ref]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}

		// call all functions that has registered for the aggregate type and ID events
		// ref also include the package name ensuring that Aggregate Types can have the same name.
		ref = formatAggregatePathNameID(aggregate)
		if subs, ok := e.specificAggregates[ref]; ok {
			for _, s := range subs {
				s.f(event)
			}
		}
	}
}

// SubscriberAll bind a function to be called on all events
func (e *EventStream) SubscriberAll(f func(e Event)) *Subscription {
	s := Subscription{
		f: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for i, sub := range e.allEvents {
			if &s == sub {
				e.allEvents = append(e.allEvents[:i], e.allEvents[i+1:]...)
				break
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		e.allEvents = append(e.allEvents, &s)
	}
	return &s
}

// SubscriberSpecificAggregate bind a function to be called on events that happen on an aggregate based on type and ID
func (e *EventStream) SubscriberSpecificAggregate(f func(e Event), aggregates ...Aggregate) *Subscription {
	s := Subscription{
		f: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			ref := formatAggregatePathNameID(a)
			for i, sub := range e.specificAggregates[ref] {
				if &s == sub {
					e.specificAggregates[ref] = append(e.specificAggregates[ref][:i], e.specificAggregates[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			ref := formatAggregatePathNameID(a)
			e.specificAggregates[ref] = append(e.specificAggregates[ref], &s)
		}
	}
	return &s
}

// SubscriberAggregateType bind a function to be called on events for an aggregate type
func (e *EventStream) SubscriberAggregateType(f func(e Event), aggregates ...Aggregate) *Subscription {
	s := Subscription{
		f: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			ref := formatAggregatePathType(a)
			for i, sub := range e.aggregateTypes[ref] {
				if &s == sub {
					e.aggregateTypes[ref] = append(e.aggregateTypes[ref][:i], e.aggregateTypes[ref][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, a := range aggregates {
			ref := formatAggregatePathType(a)
			e.aggregateTypes[ref] = append(e.aggregateTypes[ref], &s)
		}
	}
	return &s
}

// SubscriberSpecificEvent bind a function to be called on specific events
func (e *EventStream) SubscriberSpecificEvent(f func(e Event), events ...interface{}) *Subscription {
	s := Subscription{
		f: f,
	}
	s.unsubF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, event := range events {
			t := reflect.TypeOf(event)
			for i, sub := range e.specificEvents[t] {
				if &s == sub {
					e.specificEvents[t] = append(e.specificEvents[t][:i], e.specificEvents[t][i+1:]...)
					break
				}
			}
		}
	}
	s.subF = func() {
		e.lock.Lock()
		defer e.lock.Unlock()

		for _, event := range events {
			t := reflect.TypeOf(event)
			e.specificEvents[t] = append(e.specificEvents[t], &s)
		}
	}
	return &s
}

func formatAggregatePathType(a Aggregate) string {
	name := reflect.TypeOf(a).Elem().Name()
	root := a.Root()
	ref := fmt.Sprintf("%s/%s", root.path(), name)
	return ref
}

func formatAggregatePathNameID(a Aggregate) string {
	name := reflect.TypeOf(a).Elem().Name()
	root := a.Root()
	ref := fmt.Sprintf("%s/%s/%s", root.path(), name, root.ID())
	return ref
}
