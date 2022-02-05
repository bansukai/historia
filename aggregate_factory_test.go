package historia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewAggregateFactory_returns_valid_instance(t *testing.T) {
	f := NewAggregateFactory()
	assert.NotNil(t, f)
	assert.NotNil(t, f.delegates)
}

func Test_AggFactory_RegisterDelegate_should_return_error_when_delegate_returns_nil(t *testing.T) {
	f := NewAggregateFactory()
	fn := func(string) AggregateRoot { return nil }
	assert.ErrorIs(t, f.RegisterDelegate(fn), ErrDelegateFuncNotValid)
}

func Test_AggFactory_RegisterDelegate_should_return_error_when_delegate_already_registered(t *testing.T) {
	f := NewAggregateFactory()
	fn := func(string) AggregateRoot { return &factAgg{} }
	assert.NoError(t, f.RegisterDelegate(fn))
	assert.ErrorIs(t, f.RegisterDelegate(fn), ErrDelegateAlreadyRegistered)
}

func Test_AggFactory_New_should_return_error_when_aggregate_is_nil(t *testing.T) {
	f := NewAggregateFactory()
	_, err := f.New(nil, "")
	assert.ErrorIs(t, err, ErrAggregateNotValid)
}

func Test_AggFactory_New_should_return_error_when_delegate_not_registered(t *testing.T) {
	f := NewAggregateFactory()
	_, err := f.New(&factAgg{}, "")
	assert.ErrorIs(t, err, ErrDelegateNotRegistered)
}

func Test_AggFactory_New_should_return_instance(t *testing.T) {
	id := "some_id"
	f := NewAggregateFactory()
	_ = f.RegisterDelegate(func(id string) AggregateRoot { return &factAgg{AggregateBase: NewAggregateBase(id)} })
	agg, err := f.New(&factAgg{}, id)
	assert.NoError(t, err)
	assert.IsType(t, &factAgg{}, agg)
	assert.Equal(t, id, agg.ID())
}

type factAgg struct {
	AggregateBase
}

func (f *factAgg) Transition(_ Event) {}
