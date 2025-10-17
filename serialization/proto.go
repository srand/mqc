package serialization

import (
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/proto"
)

type protoSerializer struct{}

type protoEncoder struct {
	writer io.Writer
}

type protoDecoder struct {
	reader io.Reader
}

var _ Serializer = (*protoSerializer)(nil)

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
	return &protoDecoder{reader: reader}
}

func (s *protoSerializer) NewEncoder(writer io.Writer) Encoder {
	return &protoEncoder{writer: writer}
}

func (e *protoEncoder) Encode(v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidMessage
	}

	// Marshal the message
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))

	if _, err := e.writer.Write(size); err != nil {
		return err
	}

	_, err = e.writer.Write(data)
	return err
}

func (d *protoDecoder) Decode(v any) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidMessage
	}

	// Read the size prefix
	sizeBuf := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, sizeBuf); err != nil {
		return err
	}
	size := binary.LittleEndian.Uint32(sizeBuf)

	// Read the message data
	data := make([]byte, size)
	if _, err := io.ReadFull(d.reader, data); err != nil {
		return err
	}
	return proto.Unmarshal(data, msg)
}
