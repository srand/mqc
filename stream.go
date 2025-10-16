package mqc

import (
	"context"
	"io"

	"github.com/srand/mqc/serialization"
)

// ClientStreamClient represents a client-side stream for client-streaming RPCs.
// The client sends multiple requests and receives a single response.
type ClientStreamClient[Req any, Res any] interface {
	// Send a request to the server.
	// The client can send multiple requests before receiving a response.
	Send(ctx context.Context, req *Req) error

	// Close the stream and receive the final response.
	CloseAndRecv(ctx context.Context) (*Res, error)
}

// ServerStreamClient represents a client-side stream for server-streaming RPCs.
// The client sends a single request and receives multiple responses.
type ServerStreamClient[Res any] interface {
	// Recv receives a response from the server.
	// The client can call Recv multiple times to receive multiple responses.
	// When the server has finished sending responses, Recv returns io.EOF.
	Recv(ctx context.Context) (*Res, error)
}

// BidiStreamClient represents a client-side stream for bidirectional streaming RPCs.
// The client can send and receive multiple messages independently.
type BidiStreamClient[Req any, Res any] interface {
	// Send a request to the server.
	// The client can send multiple requests.
	Send(ctx context.Context, req *Req) error

	// Recv receives a response from the server.
	// The client can call Recv multiple times to receive multiple responses.
	// When the server has finished sending responses, Recv returns io.EOF.
	Recv(ctx context.Context) (*Res, error)
}

// ClientStreamServer represents a server-side stream for client-streaming RPCs.
type ClientStreamServer[Req any, Res any] interface {
	// Recv receives a request from the client.
	// The server can call Recv multiple times to receive multiple requests.
	// When the client has finished sending requests, Recv returns io.EOF.
	Recv(ctx context.Context) (*Req, error)

	// SendAndClose sends the final response to the client and closes the stream.
	SendAndClose(ctx context.Context, res *Res) error
}

// ServerStreamServer represents a server-side stream for server-streaming RPCs.
type ServerStreamServer[Res any] interface {
	// Send sends a response to the client.
	// The server can send multiple responses.
	Send(ctx context.Context, res *Res) error
}

// BidiStreamServer represents a server-side stream for bidirectional streaming RPCs.
type BidiStreamServer[Req any, Res any] interface {
	// Send sends a response to the client.
	Send(ctx context.Context, req *Res) error

	// Recv receives a request from the client.
	Recv(ctx context.Context) (*Req, error)
}

type clientStreamImpl[Req any, Res any] struct {
	call       Call
	serializer serialization.Serializer
}

func NewClientStreamClient[Req any, Res any](ctx context.Context, conn Conn, method Method) (ClientStreamClient[Req, Res], error) {
	call, err := conn.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}

	return &clientStreamImpl[Req, Res]{call: call, serializer: conn.Serializer()}, nil
}

func NewServerStreamClient[Req, Res any](ctx context.Context, conn Conn, method Method, req *Req) (ServerStreamClient[Res], error) {
	call, err := conn.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}

	msg, err := NewDataMessage(req, conn.Serializer())
	if err != nil {
		return nil, err
	}

	err = call.Send(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &clientStreamImpl[any, Res]{call: call, serializer: conn.Serializer()}, nil
}

func NewBidiStreamClient[Req any, Res any](ctx context.Context, conn Conn, method Method) (BidiStreamClient[Req, Res], error) {
	call, err := conn.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}
	return &clientStreamImpl[Req, Res]{call: call, serializer: conn.Serializer()}, nil
}

func (s *clientStreamImpl[Req, Res]) CloseAndRecv(ctx context.Context) (*Res, error) {
	// Send close signal
	if err := s.call.Send(ctx, NewCloseMessage()); err != nil {
		return nil, err
	}

	return s.Recv(ctx)
}

func (s *clientStreamImpl[Req, Res]) Recv(ctx context.Context) (*Res, error) {
	msg, err := s.call.Recv(ctx)
	if err != nil {
		return nil, err
	}

	if msg.IsError() {
		return nil, msg.Error()
	}

	if msg.IsClose() {
		return nil, io.EOF
	}

	return GetMessageData[Res](msg, s.serializer)
}

func (s *clientStreamImpl[Req, Res]) SendAndClose(ctx context.Context, res *Res) error {
	msg, err := NewDataMessage(res, s.serializer)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, msg); err != nil {
		return err
	}
	if err := s.call.Send(ctx, NewCloseMessage()); err != nil {
		return err
	}
	return nil
}

func (s *clientStreamImpl[Req, Res]) Send(ctx context.Context, req *Req) error {
	msg, err := NewDataMessage(req, s.serializer)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, msg); err != nil {
		return err
	}
	return nil
}

type serverStreamImpl[Req, Res any] struct {
	call       Call
	serializer serialization.Serializer
}

func NewClientStreamServer[Req, Res any](conn Conn, call Call) (ClientStreamServer[Req, Res], error) {
	return &serverStreamImpl[Req, Res]{call: call, serializer: conn.Serializer()}, nil
}

func NewServerStreamServer[Req, Res any](conn Conn, call Call) (ServerStreamServer[Res], *Req, error) {
	stream := &serverStreamImpl[any, Res]{call: call, serializer: conn.Serializer()}

	// ReaD initial request message
	msg, err := call.Recv(context.Background())
	if err != nil {
		return nil, nil, err
	}

	if msg.IsError() {
		return nil, nil, msg.Error()
	}

	if msg.IsClose() {
		return nil, nil, io.EOF
	}

	req, err := GetMessageData[Req](msg, conn.Serializer())
	if err != nil {
		return nil, nil, err
	}

	return stream, req, nil
}

func NewBidiStreamServer[Req, Res any](conn Conn, call Call) (BidiStreamServer[Req, Res], error) {
	return &serverStreamImpl[Req, Res]{call: call, serializer: conn.Serializer()}, nil
}

func (s *serverStreamImpl[Req, Res]) Recv(ctx context.Context) (*Req, error) {
	msg, err := s.call.Recv(ctx)
	if err != nil {
		return nil, err
	}

	if msg.IsError() {
		return nil, msg.Error()
	}

	if msg.IsClose() {
		return nil, io.EOF
	}

	return GetMessageData[Req](msg, s.serializer)
}

func (s *serverStreamImpl[Req, Res]) SendAndClose(ctx context.Context, res *Res) error {
	msg, err := NewDataMessage(res, s.serializer)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, msg); err != nil {
		return err
	}
	if err := s.call.Send(ctx, NewCloseMessage()); err != nil {
		return err
	}
	return nil
}

func (s *serverStreamImpl[Req, Res]) Send(ctx context.Context, res *Res) error {
	msg, err := NewDataMessage(res, s.serializer)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, msg); err != nil {
		return err
	}
	return nil
}
