package serialization

import (
	"encoding/gob"
	"io"

	"google.golang.org/protobuf/proto"
)

type protoSerializer struct{}

func NewProtoSerializer() Serializer {
	return &protoSerializer{}
}

func (s *protoSerializer) Marshal(v any) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, ErrInvalidMessage
	}
	return proto.Marshal(msg)
}

func (s *protoSerializer) Unmarshal(data []byte, v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidMessage
	}
	return proto.Unmarshal(data, msg)
}

func (s *protoSerializer) NewDecoder(reader io.Reader) Decoder {
	return gob.NewDecoder(reader)
}

func (s *protoSerializer) NewEncoder(writer io.Writer) Encoder {
	return gob.NewEncoder(writer)
}
