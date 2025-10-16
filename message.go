package mqc

import (
	"errors"
	"fmt"

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

func NewCallMessage(method Method) *Message {
	return &Message{
		Type: MsgTypeCall,
		Data: []byte(method),
	}
}

func NewDataMessage[Req any](req *Req, serializer serialization.Serializer) (*Message, error) {
	data, err := serializer.Marshal(req)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type: MsgTypeData,
		Data: data,
	}, nil
}

func NewErrorMessage(err error, serializer serialization.Serializer) *Message {
	println("NewErrorMessage called", err.Error()) // --- IGNORE ---
	return &Message{
		Type: MsgTypeError,
		Data: []byte(err.Error()),
	}
}

func NewCloseMessage() *Message {
	return &Message{
		Type: MsgTypeClose,
	}
}

func NewAckMessage() *Message {
	return &Message{
		Type: MsgTypeAck,
	}
}

func (m *Message) IsAck() bool {
	return m.Type == MsgTypeAck
}

func (m *Message) IsCall() bool {
	return m.Type == MsgTypeCall
}

func (m *Message) IsData() bool {
	return m.Type == MsgTypeData
}

func (m *Message) IsError() bool {
	return m.Type == MsgTypeError
}

func (m *Message) IsClose() bool {
	return m.Type == MsgTypeClose
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

func GetMessageData[T any](msg *Message, serializer serialization.Serializer) (*T, error) {
	if !msg.IsData() {
		return nil, fmt.Errorf("message is not a data message: %v+", msg)
	}
	var obj T
	if err := serializer.Unmarshal(msg.Data, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}
