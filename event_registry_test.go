package historia

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewEventRegistry_should_return_valid_instance(t *testing.T) {
	registry := NewEventRegistry()
	assert.NotNil(t, registry.eventDataNameFormatter)
	assert.NotNil(t, registry.factories)

	fn := func(data EventData) string { return "yay" }

	registry = NewEventRegistry(WithEventDataNameFormatter(fn))
	assert.Equal(t, reflect.ValueOf(fn).Pointer(), reflect.ValueOf(registry.eventDataNameFormatter).Pointer())
}

func Test_EventRegistry_Register(t *testing.T) {
	t.Run("factory should return pointer", func(t *testing.T) {
		registry := NewEventRegistry()
		err := registry.Register(factoryBadFn)
		assert.ErrorIs(t, err, ErrFactoryShouldReturnValidPointer)

		err = registry.Register(factoryNilFn)
		assert.ErrorIs(t, err, ErrFactoryShouldReturnValidPointer)
	})

	t.Run("factory should return proper type", func(t *testing.T) {
		registry := NewEventRegistry()
		err := registry.Register(factoryFn)
		assert.NoError(t, err)
		assert.Len(t, registry.factories, 1)

		td, err := registry.Create(registry.GetName(&eventRegistryTestData{}))
		assert.NoError(t, err)
		assert.IsType(t, &eventRegistryTestData{}, td)
	})

	t.Run("should not register duplicate factories", func(t *testing.T) {
		registry := NewEventRegistry()
		_ = registry.Register(factoryFn)
		err := registry.Register(factoryFn)
		assert.NoError(t, err)
		assert.Len(t, registry.factories, 1)
	})
}

func Test_EventRegistry_Unregister(t *testing.T) {
	t.Run("should return early if type not registered", func(t *testing.T) {
		registry := NewEventRegistry()
		registry.Unregister(&eventRegistryTestData{})
		assert.Len(t, registry.factories, 0)
	})

	t.Run("should remove factory", func(t *testing.T) {
		registry := NewEventRegistry()
		_ = registry.Register(factoryFn)
		registry.Unregister(factoryFn())
		assert.Len(t, registry.factories, 0)
	})
}

func Test_EventRegistry_Create(t *testing.T) {
	t.Run("should return error when type not registered", func(t *testing.T) {
		registry := NewEventRegistry()
		_, err := registry.Create("poo")
		assert.ErrorIs(t, err, ErrEventDataFactoryNotRegistered)
	})

	t.Run("should return pointer", func(t *testing.T) {
		registry := NewEventRegistry()
		_ = registry.Register(factoryFn)
		ptr, err := registry.Create(registry.GetName(factoryFn()))
		assert.NoError(t, err)
		assert.IsType(t, &eventRegistryTestData{}, ptr)
	})
}

func Test_dn(t *testing.T) {
	n1 := defaultNameFormatter(eventRegistryTestData{})
	n2 := defaultNameFormatter(&eventRegistryTestData{})

	fmt.Println(n1)
	fmt.Println(n2)
}

func factoryBadFn() EventData { return eventRegistryTestData{} }
func factoryNilFn() EventData { return nil }
func factoryFn() EventData    { return &eventRegistryTestData{} }

type eventRegistryTestData struct {
	Something string
	Other     int
}
