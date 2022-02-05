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

type AggregateRoot interface {
	ID() string
	Version() Version
	Events() []Event
}

// AggregateFactory returns aggregate instances of a specified type with the
// AggregateID set to the provided value.
type AggregateFactory interface {
	New(aggregate AggregateRoot, id string) (AggregateRoot, error)
}

// NewAggregateFactory returns a new AggFactory
func NewAggregateFactory() *AggFactory {
	return &AggFactory{
		delegates: make(map[string]func(string) AggregateRoot),
	}
}

// AggFactory is an implementation of the AggregateFactory interface
// that supports registration of delegate functions to perform aggregate instantiation.
type AggFactory struct {
	delegates map[string]func(string) AggregateRoot
}

// RegisterDelegate is used to register a new function for instantiation of an aggregate.
//
// Examples:
// 	func(id string) AggregateBase { return NewMyAggregateType(id) }
// 	func(id string) AggregateBase { return &MyAggregateType{AggregateBase:NewAggregateBase(id)} }
func (f *AggFactory) RegisterDelegate(delegate func(string) AggregateRoot) error {
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

// New calls the delegate for the type specified and returns the result.
func (f *AggFactory) New(aggregate AggregateRoot, id string) (AggregateRoot, error) {
	if aggregate == nil {
		return nil, ErrAggregateNotValid
	}

	name := TypeOf(aggregate)
	delegate, ok := f.delegates[name]
	if !ok {
		return nil, ErrDelegateNotRegistered
	}

	return delegate(id), nil
}
