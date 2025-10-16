package conn

import (
	"context"
	"net"

	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
)

// Length-prefixed call
type call struct {
	conn    net.Conn
	decoder serialization.Decoder
	encoder serialization.Encoder
}

func NewCall(conn net.Conn, serializer serialization.Serializer) *call {
	return &call{
		conn:    conn,
		decoder: serializer.NewDecoder(conn),
		encoder: serializer.NewEncoder(conn),
	}
}

func (s *call) Close() error {
	return s.conn.Close()
}

func (s *call) Send(ctx context.Context, msg *mqc.Message) error {
	return s.encoder.Encode(msg)
}

func (s *call) Recv(ctx context.Context) (*mqc.Message, error) {
	var msg mqc.Message

	if err := s.decoder.Decode(&msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
