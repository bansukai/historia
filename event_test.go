package historia

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Reason_should_return_name_of_event(t *testing.T) {
	evt := Event{}
	assert.Equal(t, "", evt.Reason())

	type moo struct{}
	evt.Data = &moo{}
	assert.Equal(t, "moo", evt.Reason())
}
