package historia

import (
	"errors"
)

var (
	ErrDelegateAlreadyRegistered = errors.New("delegate already registered")
	ErrDelegateNotRegistered     = errors.New("no delegate registered for this aggregate")
	ErrDelegateFuncNotValid      = errors.New("delegate func didn't return a valid aggregate")
	ErrAggregateNotValid         = errors.New("aggregate is nil")
)

// AggregateFactory returns aggregate instances of a specified type with the
// AggregateID set to the provided value.
type AggregateFactory interface {
	Get(aggregate Aggregate, id string) (Aggregate, error)
}

// AggregateRepository defines a contract for loading aggregates from the event history
type AggregateRepository interface {
	// Get fetches the aggregates event and builds up the aggregate
	// If there is a snapshot store, try to fetch a snapshot of the aggregate and
	// event after the version of the aggregate, if any.
	Get(id string, aggregate Aggregate) error
}

// NewAggregateFactory returns a new AggFactory
func NewAggregateFactory(repository AggregateRepository) *AggFactory {
	return &AggFactory{
		repository: repository,
		delegates:  make(map[string]func(string) Aggregate),
	}
}

// AggFactory is an implementation of the AggregateFactory interface
// that supports registration of delegate functions to perform aggregate instantiation.
type AggFactory struct {
	repository AggregateRepository
	delegates  map[string]func(string) Aggregate
}

// RegisterDelegate is used to register a new function for instantiation of an aggregate.
//
// Examples:
// 	func(id string) AggregateBase { return NewMyAggregateType(id) }
// 	func(id string) AggregateBase { return &MyAggregateType{AggregateBase:NewAggregateBase(id)} }
func (f *AggFactory) RegisterDelegate(delegate func(string) Aggregate) error {
	aggregate := delegate("_not_an_id_")
	if aggregate == nil {
		return ErrDelegateFuncNotValid
	}

	typeName := TypeOf(aggregate)
	if _, ok := f.delegates[typeName]; ok {
		return ErrDelegateAlreadyRegistered
	}

	f.delegates[typeName] = delegate
	return nil
}

// Get uses the registered delegate for the Aggregate and attempts to load
// its history from the AggregateRepository.
func (f *AggFactory) Get(agType Aggregate, id string) (Aggregate, error) {
	if agType == nil {
		return nil, ErrAggregateNotValid
	}

	name := TypeOf(agType)
	delegate, ok := f.delegates[name]
	if !ok {
		return nil, ErrDelegateNotRegistered
	}

	aggregate := delegate(id)
	if err := f.repository.Get(id, aggregate); err != nil {
		if !errors.Is(err, ErrAggregateNotFound) {
			return nil, err
		}
	}

	return aggregate, nil
}
