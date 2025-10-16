package conn

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
)

// Length-prefixed call
type call struct {
	conn       net.Conn
	decoder    serialization.Decoder
	encoder    serialization.Encoder
	receiver   chan *mqc.Message
	serializer serialization.Serializer
	err        error
}

func NewCall(conn net.Conn, serializer serialization.Serializer) *call {
	call := &call{
		conn:       conn,
		decoder:    serializer.NewDecoder(conn),
		encoder:    serializer.NewEncoder(conn),
		receiver:   make(chan *mqc.Message),
		serializer: serializer,
	}
	go call.run()
	return call
}

func (s *call) Close() error {
	return s.conn.Close()
}

func (s *call) Send(ctx context.Context, msg *mqc.Message) error {
	if msg == nil {
		return errors.New("message is nil")
	}
	if s.err != nil {
		return s.err
	}
	return s.encoder.Encode(msg)
}

func (s *call) Recv(ctx context.Context) (*mqc.Message, error) {
	var msg *mqc.Message

	if s.err != nil {
		return nil, s.err
	}

	select {
	case msg = <-s.receiver:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if msg == nil {
		return nil, io.EOF
	}

	if msg.IsError() {
		s.err = msg.Error()
		return nil, s.err
	}

	return msg, nil
}

func (c *call) run() {
	defer close(c.receiver)
	for {
		var msg mqc.Message

		if err := c.decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			// Connection closed or error occurred
			c.receiver <- mqc.NewErrorMessage(err, c.serializer)
			return
		}

		if msg.IsError() {
			c.err = msg.Error()
			c.receiver <- &msg
			return
		}

		c.receiver <- &msg

	}
}
