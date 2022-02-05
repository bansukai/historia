package snapshot

import (
	"context"
	"testing"

	"github.com/bansukai/historia"
	"github.com/stretchr/testify/assert"
)

func AcceptanceTestSnapshotStore(t *testing.T, snapshot historia.SnapshotStore) {
	expected := historia.Snapshot{
		Version: 10,
		ID:      "123",
		Type:    "Person",
		State:   []byte{},
	}

	assert.NoError(t, snapshot.Save(context.Background(), &expected))

	_, err := snapshot.Get(context.Background(), "bogus", "bogus")
	assert.ErrorIs(t, err, historia.ErrSnapshotNotFound)

	actual, err := snapshot.Get(context.Background(), "123", "Person")

	assert.NoError(t, err)
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Type, actual.Type)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.Version, actual.Version)
	assert.Equal(t, expected.State, actual.State)
}
