package mqc

import (
	"errors"

	"github.com/srand/mqc/serialization"
)

const (
	// MsgTypeCall represents a call message.
	MsgTypeCall byte = 0x00
	// MsgTypeData represents a data message.
	MsgTypeData byte = 0x01
	// MsgTypeError represents an error message.
	MsgTypeError byte = 0x02
	// MsgTypeClose represents a close message.
	MsgTypeClose byte = 0x03
	// MsgTypeAck represents an acknowledgment message.
	MsgTypeAck byte = 0x80
)

// Message represents a control message in the mqc package.
type Message struct {
	// Type of the message (e.g., call, data, error, close, ack).
	Type byte

	// Data associated with the message.
	// For Call messages, this is the method name.
	// For Data messages, this is the serialized payload.
	// For Error messages, this is the error string.
	// For Close and Ack messages, this is nil.
	Data []byte
}

func NewAckMessage() *Message {
	return &Message{
		Type: MsgTypeAck,
	}
}
func NewCallMessage(method Method) *Message {
	return &Message{
		Type: MsgTypeCall,
		Data: []byte(method),
	}
}

func NewCloseMessage() *Message {
	return &Message{
		Type: MsgTypeClose,
	}
}

func NewDataMessage(data []byte) *Message {
	return &Message{
		Type: MsgTypeData,
		Data: data,
	}
}

func NewErrorMessage(err error) *Message {
	return &Message{
		Type: MsgTypeError,
		Data: []byte(err.Error()),
	}
}

func (m *Message) IsAck() bool {
	return m.Type == MsgTypeAck
}

func (m *Message) IsCall() bool {
	return m.Type == MsgTypeCall
}

func (m *Message) IsClose() bool {
	return m.Type == MsgTypeClose
}

func (m *Message) IsData() bool {
	return m.Type == MsgTypeData
}

func (m *Message) IsError() bool {
	return m.Type == MsgTypeError
}

func (m *Message) Error() error {
	if m.IsError() {
		return errors.New(string(m.Data))
	}
	return nil
}

func (m *Message) Method() Method {
	return Method(string(m.Data))
}

func (m *Message) DataBytes() []byte {
	return m.Data
}

func marshal[T any](serializer serialization.Serializer, data *T) ([]byte, error) {
	bytes, err := serializer.Marshal(data)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func unmarshal[T any](serializer serialization.Serializer, data []byte) (*T, error) {
	var result T
	err := serializer.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
