package common

import (
	"context"
	"errors"
	"io"
	"net"

	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
)

// Represents a call connection over net.Conn transport
type callConn struct {
	conn       net.Conn
	decoder    serialization.Decoder
	encoder    serialization.Encoder
	receiver   chan *mqc.Message
	serializer serialization.Serializer
	err        error
}

var _ mqc.Conn = (*callConn)(nil)

func NewConn(conn net.Conn, serializer serialization.Serializer) *callConn {
	cc := &callConn{
		conn:       conn,
		decoder:    serializer.NewDecoder(conn),
		encoder:    serializer.NewEncoder(conn),
		receiver:   make(chan *mqc.Message),
		serializer: serializer,
	}
	go cc.run()
	return cc
}

func (s *callConn) Close() error {
	return s.conn.Close()
}

func (s *callConn) Send(ctx context.Context, data []byte) error {
	if data == nil {
		return errors.New("data is nil")
	}
	if s.err != nil {
		return s.err
	}

	return s.encoder.Encode(mqc.NewDataMessage(data))
}

func (s *callConn) sendControl(ctx context.Context, msg *mqc.Message) error {
	if s.err != nil {
		return s.err
	}

	return s.encoder.Encode(msg)
}

func (s *callConn) SendAck(ctx context.Context) error {
	return s.sendControl(ctx, mqc.NewAckMessage())
}

func (s *callConn) SendClose(ctx context.Context) error {
	return s.sendControl(ctx, mqc.NewCloseMessage())
}

func (s *callConn) SendError(ctx context.Context, err error) error {
	return s.sendControl(ctx, mqc.NewErrorMessage(err))
}

func (s *callConn) SendMethod(ctx context.Context, method mqc.Method) error {
	return s.sendControl(ctx, mqc.NewCallMessage(method))
}

func (s *callConn) Recv(ctx context.Context) ([]byte, error) {
	var msg *mqc.Message

	if s.err != nil {
		return nil, s.err
	}

	select {
	case msg = <-s.receiver:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	if msg == nil || msg.IsClose() {
		return nil, io.EOF
	}

	if msg.IsError() {
		s.err = msg.Error()
		return nil, s.err
	}

	if !msg.IsData() {
		return nil, mqc.ErrProtocolViolation
	}

	return msg.DataBytes(), nil
}

func (s *callConn) RecvMethod(ctx context.Context) (mqc.Method, error) {
	var msg *mqc.Message

	if s.err != nil {
		return "", s.err
	}

	select {
	case msg = <-s.receiver:
	case <-ctx.Done():
		return "", ctx.Err()
	}

	if msg == nil || msg.IsClose() {
		return "", io.EOF
	}

	if msg.IsError() {
		s.err = msg.Error()
		return "", s.err
	}

	if !msg.IsCall() {
		return "", mqc.ErrProtocolViolation
	}

	return msg.Method(), nil
}

func (c *callConn) run() {
	defer close(c.receiver)
	for {
		var msg mqc.Message

		if err := c.decoder.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			// Connection closed or error occurred
			c.receiver <- mqc.NewErrorMessage(err)
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
