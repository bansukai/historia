package historia

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

var (
	ErrFactoryShouldReturnValidPointer = errors.New("the provided factory didn't return a valid pointer")
	ErrEventDataFactoryNotRegistered   = errors.New("event data factory not registered for this type")

	DefaultRegistry = NewEventRegistry()
)

type EventRegistry interface {
	// Register registers an event data factory for a type. The factory is
	// used to create concrete event data structs when loading from a database.
	Register(factory func() EventData) error

	// Unregister removes the registration of the event data factory for
	// a type. This is mainly useful in maintenance situations where the event data
	// needs to be switched in a migrations.
	Unregister(data EventData)

	// Create creates an event data of a type using the factory registered
	// with Register.
	Create(name string) (EventData, error)

	// GetName returns the name assigned to the EventData type.
	GetName(data EventData) string
}

type Option func(registry *EventRegister)

func WithEventDataNameFormatter(fn func(data EventData) string) Option {
	return func(registry *EventRegister) {
		registry.eventDataNameFormatter = fn
	}
}

func NewEventRegistry(opts ...Option) *EventRegister {
	er := &EventRegister{
		factories:              map[string]func() EventData{},
		eventDataNameFormatter: defaultNameFormatter,
	}

	for _, opt := range opts {
		opt(er)
	}

	return er
}

// RegisterAllOf uses the Register method on the DefaultRegistry to register all factories.
// note: this method will panic if there is an error registering a factory item.
func RegisterAllOf(items []EventData) {
	for i := range items {
		e := items[i]
		fn := func() EventData { return e }

		if err := DefaultRegistry.Register(fn); err != nil {
			panic(err.Error())
		}
	}
}

// RegisterEventData uses the DefaultRegistry's Register method to register the factory
func RegisterEventData(factory func() EventData) error {
	return DefaultRegistry.Register(factory)
}

// UnregisterEventData uses the DefaultRegistry's Unregister method to unregister the factory
func UnregisterEventData(data EventData) {
	DefaultRegistry.Unregister(data)
}

// Create uses the DefaultRegistry's Create method to create the EventData
func Create(name string) (EventData, error) {
	return DefaultRegistry.Create(name)
}

type EventRegister struct {
	factories              map[string]func() EventData
	mu                     sync.RWMutex
	eventDataNameFormatter func(data EventData) string
}

func (e *EventRegister) Register(factory func() EventData) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	data := factory()

	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrFactoryShouldReturnValidPointer
	}

	name := e.GetName(data)
	if _, ok := e.factories[name]; ok {
		return nil
	}

	e.factories[name] = factory
	return nil
}

func (e *EventRegister) Unregister(data EventData) {
	e.mu.Lock()
	defer e.mu.Unlock()

	name := e.GetName(data)
	if _, ok := e.factories[name]; !ok {
		return
	}

	delete(e.factories, name)
}

func (e *EventRegister) Create(name string) (EventData, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if factory, ok := e.factories[name]; ok {
		return factory(), nil
	}

	return nil, ErrEventDataFactoryNotRegistered
}

func (e *EventRegister) GetName(data EventData) string {
	return e.eventDataNameFormatter(data)
}

func defaultNameFormatter(data EventData) string {
	to := reflect.TypeOf(data)
	for to.Kind() == reflect.Ptr {
		to = to.Elem()
	}

	return fmt.Sprintf("%s#%s", to.PkgPath(), to.Name())
}
