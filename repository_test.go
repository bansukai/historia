package historia

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRepository_should_return_valid_instance(t *testing.T) {
	esMock := &eventStoreMocker{}
	snapMock := &snapMocker{}
	repo := NewRepository(esMock, snapMock)

	assert.Equal(t, esMock, repo.eventStore)
	assert.Equal(t, snapMock, repo.snapper)
}

func Test_Repo_Save_should_return_eventStore_error(t *testing.T) {
	e := errors.New("it went boom")
	esMock := &eventStoreMocker{
		save: func(context.Context, []Event) error { return e },
	}
	snapMock := &snapMocker{}
	repo := NewRepository(esMock, snapMock)

	p1 := repoAggregate{}
	_ = p1.SetID("hi_there")
	p1.TrackChange(&p1, &repoEvent1{})
	assert.ErrorIs(t, repo.Save(context.Background(), &p1), e)
}

func Test_Repo_Save_should_call_eventStore_Update(t *testing.T) {
	esMock := &eventStoreMocker{
		save: func(context.Context, []Event) error { return nil },
	}
	snapMock := &snapMocker{}
	eventFired := false
	repo := NewRepository(esMock, snapMock)
	repo.SubscriberAll(func(ctx context.Context, e Event) error {
		eventFired = true
		return nil
	}).Subscribe()

	p1 := repoAggregate{}
	_ = p1.SetID("hi_there")
	p1.TrackChange(&p1, &repoEvent1{})
	assert.NoError(t, repo.Save(context.Background(), &p1))
	assert.True(t, eventFired)

	assert.Equal(t, Version(1), p1.Version())
	assert.Len(t, p1.Events(), 0)
}

func Test_Repo_Get_should_return_error_when_snapper_fails(t *testing.T) {
	err := errors.New("well no")
	sn := &snapMocker{
		apply: func(context.Context, string, Aggregate) error { return err },
	}
	repo := NewRepository(nil, sn)
	assert.ErrorIs(t, repo.Get(context.Background(), "asd", &repoAggregate{}), err)
}

func Test_Repo_Get_should_return_error_when_eventStore_fails(t *testing.T) {
	err := errors.New("well no")
	es := &eventStoreMocker{
		get: func(context.Context, string, string, Version) ([]Event, error) {
			return nil, err
		},
	}
	sn := &snapMocker{
		apply: func(context.Context, string, Aggregate) error { return nil },
	}
	repo := NewRepository(es, sn)
	assert.ErrorIs(t, repo.Get(context.Background(), "asd", &repoAggregate{}), err)
}

func Test_Repo_Get_should_return_error_when_eventStore_has_no_events_but_root_is_version_zero(t *testing.T) {
	es := &eventStoreMocker{
		get: func(context.Context, string, string, Version) ([]Event, error) {
			return nil, ErrNoEvents
		},
	}
	sn := &snapMocker{
		apply: func(context.Context, string, Aggregate) error { return nil },
	}
	repo := NewRepository(es, sn)
	assert.ErrorIs(t, repo.Get(context.Background(), "asd", &repoAggregate{}), ErrAggregateNotFound)
}

func Test_Repo_Get_should_build_aggregate_from_history(t *testing.T) {
	events := []Event{
		{Version: 1, Data: &repoEvent1{Name: "Poo"}, AggregateType: "repoAggregate"},
		{Version: 2, Data: &repoEvent1{Name: "Happy"}, AggregateType: "repoAggregate"},
	}

	es := &eventStoreMocker{
		get: func(context.Context, string, string, Version) ([]Event, error) {
			return events, nil
		},
	}

	sn := &snapMocker{
		apply: func(context.Context, string, Aggregate) error { return nil },
	}

	repo := NewRepository(es, sn)
	ag := &repoAggregate{}
	assert.NoError(t, repo.Get(context.Background(), "asd", ag))

	assert.Equal(t, events[1].Version, ag.Version())
	assert.Equal(t, "Happy", ag.name)
}

func Test_Repo_SaveSnapshot_should_return_error_if_no_snapper(t *testing.T) {
	r := NewRepository(nil, nil)
	assert.ErrorIs(t, r.SaveSnapshot(context.Background(), &repoAggregate{}), ErrNoSnapShotInitialized)
}

func Test_Repo_SaveSnapshot_should_call_snapshot_save(t *testing.T) {
	err := errors.New("failed")
	sn := &snapMocker{
		save: func(ctx context.Context, aggregate Aggregate) error { return err },
	}
	r := NewRepository(nil, sn)
	assert.ErrorIs(t, r.SaveSnapshot(context.Background(), &repoAggregate{}), err)
}

// region Mocks

type repoAggregate struct {
	AggregateBase
	name string
}

func (r *repoAggregate) Transition(evt Event) {
	switch e := evt.Data.(type) {
	case *repoEvent1:
		r.name = e.Name
	default:

	}
}

type repoEvent1 struct {
	Name string
}

type eventStoreMocker struct {
	save  func(ctx context.Context, events []Event) error
	get   func(ctx context.Context, aggregateID string, aggregateType string, afterVersion Version) ([]Event, error)
	close func() error
}

func (e *eventStoreMocker) SaveEvents(ctx context.Context, events []Event) error {
	return e.save(ctx, events)
}

func (e *eventStoreMocker) GetEvents(ctx context.Context, aggregateID string, aggregateType string, afterVersion Version) ([]Event, error) {
	return e.get(ctx, aggregateID, aggregateType, afterVersion)
}

func (e *eventStoreMocker) Close() error {
	return e.close()
}

type snapMocker struct {
	apply func(ctx context.Context, aggregateID string, aggregate Aggregate) error
	save  func(ctx context.Context, aggregate Aggregate) error
}

func (s *snapMocker) ApplySnapshot(ctx context.Context, aggregateID string, aggregate Aggregate) error {
	return s.apply(ctx, aggregateID, aggregate)
}

func (s *snapMocker) SaveSnapshot(ctx context.Context, aggregate Aggregate) error {
	return s.save(ctx, aggregate)
}

// endregion
