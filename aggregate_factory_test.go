package historia

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewAggregateFactory_returns_valid_instance(t *testing.T) {
	f := NewAggregateFactory(nil)
	assert.NotNil(t, f)
	assert.NotNil(t, f.delegates)
}

func Test_AggFactory_RegisterDelegate_should_return_error_when_delegate_returns_nil(t *testing.T) {
	f := NewAggregateFactory(nil)
	fn := func(string) Aggregate { return nil }
	assert.ErrorIs(t, f.RegisterDelegate(fn), ErrDelegateFuncNotValid)
}

func Test_AggFactory_RegisterDelegate_should_return_error_when_delegate_already_registered(t *testing.T) {
	f := NewAggregateFactory(nil)
	fn := func(string) Aggregate { return &factAgg{} }
	assert.NoError(t, f.RegisterDelegate(fn))
	assert.ErrorIs(t, f.RegisterDelegate(fn), ErrDelegateAlreadyRegistered)
}

func Test_AggFactory_Get_should_return_error_when_aggregate_is_nil(t *testing.T) {
	f := NewAggregateFactory(nil)
	_, err := f.Get(nil, "")
	assert.ErrorIs(t, err, ErrAggregateNotValid)
}

func Test_AggFactory_Get_should_return_error_when_delegate_not_registered(t *testing.T) {
	f := NewAggregateFactory(nil)
	_, err := f.Get(&factAgg{}, "")
	assert.ErrorIs(t, err, ErrDelegateNotRegistered)
}

func Test_AggFactory_Get_should_return_error_when_factory_get_returns_error(t *testing.T) {
	id := "some_id"
	eErr := errors.New("it exploded")

	mock := &agRepoMocker{
		get: func(id string, aggregate Aggregate) error {
			return eErr
		},
	}

	f := NewAggregateFactory(mock)
	_ = f.RegisterDelegate(func(id string) Aggregate { return &factAgg{AggregateBase: NewAggregateBase(id)} })
	_, err := f.Get(&factAgg{}, id)
	assert.ErrorIs(t, err, eErr)
}

func Test_AggFactory_Get_should_return_instance_when_factory_returns_not_found(t *testing.T) {
	id := "some_id"
	mock := &agRepoMocker{
		get: func(id string, aggregate Aggregate) error {
			return ErrAggregateNotFound
		},
	}

	f := NewAggregateFactory(mock)
	_ = f.RegisterDelegate(func(id string) Aggregate { return &factAgg{AggregateBase: NewAggregateBase(id)} })
	agg, err := f.Get(&factAgg{}, id)
	assert.NoError(t, err)
	assert.IsType(t, &factAgg{}, agg)
	assert.Equal(t, id, agg.Root().ID())
}

func Test_AggFactory_Get_should_return_instance(t *testing.T) {
	id := "some_id"
	getCalled := false

	mock := &agRepoMocker{
		get: func(id string, aggregate Aggregate) error {
			getCalled = true
			return nil
		},
	}

	f := NewAggregateFactory(mock)
	_ = f.RegisterDelegate(func(id string) Aggregate { return &factAgg{AggregateBase: NewAggregateBase(id)} })
	agg, err := f.Get(&factAgg{}, id)
	assert.True(t, getCalled)
	assert.NoError(t, err)
	assert.IsType(t, &factAgg{}, agg)
	assert.Equal(t, id, agg.Root().ID())
}

type factAgg struct {
	AggregateBase
}

func (f *factAgg) Transition(_ Event) {}

type agRepoMocker struct {
	get func(id string, aggregate Aggregate) error
}

func (a *agRepoMocker) Get(id string, aggregate Aggregate) error {
	return a.get(id, aggregate)
}
