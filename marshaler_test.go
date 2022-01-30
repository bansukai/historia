package historia

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewMarshal_should_return_instance(t *testing.T) {
	m := NewJSONMarshal()
	assert.Equal(t, reflect.ValueOf(json.Marshal).Pointer(), reflect.ValueOf(m.marshal).Pointer())
	assert.Equal(t, reflect.ValueOf(json.Unmarshal).Pointer(), reflect.ValueOf(m.unmarshal).Pointer())
}

func Test_Marshal_marshal(t *testing.T) {
	m := NewJSONMarshal()
	d1 := marData1{A: 123, B: "asd"}
	buff, err := m.Marshal(d1)
	expected, _ := json.Marshal(d1)
	assert.NoError(t, err)
	assert.Equal(t, expected, buff)
}

func Test_Marshal_unmarshal(t *testing.T) {
	m := NewJSONMarshal()

	d1 := marData1{A: 123, B: "asd"}
	buff, _ := json.Marshal(d1)
	var actual marData1
	assert.NoError(t, m.Unmarshal(buff, &actual))
	assert.Equal(t, d1, actual)
}

type marData1 struct {
	A int
	B string
}
