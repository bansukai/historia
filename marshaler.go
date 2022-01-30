package historia

import (
	"encoding/json"
)

type marshal func(v interface{}) ([]byte, error)
type unmarshal func(data []byte, v interface{}) error

func NewJSONMarshal() *Marshal {
	return NewMarshal(json.Marshal, json.Unmarshal)
}

func NewMarshal(marshalF marshal, unmarshalF unmarshal) *Marshal {
	return &Marshal{
		marshal:   marshalF,
		unmarshal: unmarshalF,
	}
}

type Marshal struct {
	marshal   marshal
	unmarshal unmarshal
}

// Marshal pass the request to the under laying Marshal method
func (h *Marshal) Marshal(v interface{}) ([]byte, error) {
	return h.marshal(v)
}

// Unmarshal pass the request to the under laying Unmarshal method
func (h *Marshal) Unmarshal(data []byte, v interface{}) error {
	return h.unmarshal(data, v)
}
