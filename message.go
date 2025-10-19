//go:generate protoc -I. --go_out=.. message.proto
package mqc

import (
	"errors"

	"github.com/srand/mqc/serialization"
)

func NewAckMessage() *Message {
	return &Message{
		Type: Message_ACK,
	}
}
func NewCallMessage(method *Method) *Message {
	return &Message{
		Type: Message_INVOKE,
		Data: []byte(method.String()),
	}
}

func NewCloseMessage() *Message {
	return &Message{
		Type: Message_CLOSE,
	}
}

func NewDataMessage(data []byte) *Message {
	return &Message{
		Type: Message_DATA,
		Data: data,
	}
}

func NewErrorMessage(err error) *Message {
	return &Message{
		Type: Message_ERROR,
		Data: []byte(err.Error()),
	}
}

func (m *Message) IsAck() bool {
	return m.Type == Message_ACK
}

func (m *Message) IsCall() bool {
	return m.Type == Message_INVOKE
}

func (m *Message) IsClose() bool {
	return m.Type == Message_CLOSE
}

func (m *Message) IsData() bool {
	return m.Type == Message_DATA
}

func (m *Message) IsError() bool {
	return m.Type == Message_ERROR
}

func (m *Message) Error() error {
	if m.IsError() {
		return errors.New(string(m.Data))
	}
	return nil
}

func (m *Message) Method() *Method {
	return NewMethodFromString(string(m.Data))
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
