package mqc

import (
	"context"
	"errors"
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

	// Close the stream.
	CloseSend() error
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

	// CloseSend closes the stream for sending.
	CloseSend() error
}

type clientStreamImpl[Req any, Res any] struct {
	call       Conn
	serializer serialization.Serializer
	eof        bool
}

func NewClientStreamClient[Req any, Res any](ctx context.Context, transport Transport, method *Method) (ClientStreamClient[Req, Res], error) {
	call, err := transport.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}

	return &clientStreamImpl[Req, Res]{call: call, serializer: transport.Serializer()}, nil
}

func NewServerStreamClient[Req, Res any](ctx context.Context, transport Transport, method *Method, req *Req) (ServerStreamClient[Res], error) {
	serializer := transport.Serializer()

	call, err := transport.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}

	data, err := serializer.Marshal(req)
	if err != nil {
		return nil, err
	}

	err = call.Send(ctx, data)
	if err != nil {
		return nil, err
	}

	return &clientStreamImpl[any, Res]{call: call, serializer: serializer}, nil
}

func NewBidiStreamClient[Req any, Res any](ctx context.Context, transport Transport, method *Method) (BidiStreamClient[Req, Res], error) {
	call, err := transport.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}
	return &clientStreamImpl[Req, Res]{call: call, serializer: transport.Serializer()}, nil
}

func (s *clientStreamImpl[Req, Res]) CloseAndRecv(ctx context.Context) (*Res, error) {
	// Send close signal
	if err := s.call.SendClose(ctx); err != nil {
		return nil, err
	}

	return s.Recv(ctx)
}

func (s *clientStreamImpl[Req, Res]) Recv(ctx context.Context) (*Res, error) {
	if s.eof {
		return nil, io.EOF
	}

	data, err := s.call.Recv(ctx)
	if errors.Is(err, io.EOF) {
		s.eof = true
		return nil, io.EOF
	}

	if err != nil {
		return nil, err
	}

	return unmarshal[Res](s.serializer, data)
}

func (s *clientStreamImpl[Req, Res]) SendAndClose(ctx context.Context, res *Res) error {
	data, err := marshal[Res](s.serializer, res)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, data); err != nil {
		return err
	}
	if err := s.call.SendClose(ctx); err != nil {
		return err
	}
	return nil
}

func (s *clientStreamImpl[Req, Res]) Send(ctx context.Context, req *Req) error {
	data, err := marshal[Req](s.serializer, req)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, data); err != nil {
		return err
	}
	return nil
}

func (s *clientStreamImpl[Req, Res]) CloseSend() error {
	return s.call.SendClose(context.Background())
}

type serverStreamImpl[Req, Res any] struct {
	call       Conn
	serializer serialization.Serializer
	eof        bool
}

func NewClientStreamServer[Req, Res any](transport Transport, call Conn) (ClientStreamServer[Req, Res], error) {
	return &serverStreamImpl[Req, Res]{call: call, serializer: transport.Serializer()}, nil
}

func NewServerStreamServer[Req, Res any](transport Transport, call Conn) (ServerStreamServer[Res], *Req, error) {
	stream := &serverStreamImpl[any, Res]{call: call, serializer: transport.Serializer()}

	// Read initial request message
	data, err := call.Recv(context.Background())
	if err != nil {
		return nil, nil, err
	}

	req, err := unmarshal[Req](transport.Serializer(), data)
	if err != nil {
		return nil, nil, err
	}

	return stream, req, nil
}

func NewBidiStreamServer[Req, Res any](transport Transport, call Conn) (BidiStreamServer[Req, Res], error) {
	return &serverStreamImpl[Req, Res]{call: call, serializer: transport.Serializer()}, nil
}

func (s *serverStreamImpl[Req, Res]) Recv(ctx context.Context) (*Req, error) {
	if s.eof {
		return nil, io.EOF
	}

	data, err := s.call.Recv(ctx)
	if errors.Is(err, io.EOF) {
		s.eof = true
		return nil, io.EOF
	}
	if err != nil {
		return nil, err
	}

	return unmarshal[Req](s.serializer, data)
}

func (s *serverStreamImpl[Req, Res]) SendAndClose(ctx context.Context, res *Res) error {
	data, err := marshal[Res](s.serializer, res)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, data); err != nil {
		return err
	}
	if err := s.call.SendClose(ctx); err != nil {
		return err
	}
	return nil
}

func (s *serverStreamImpl[Req, Res]) Send(ctx context.Context, res *Res) error {
	data, err := marshal[Res](s.serializer, res)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, data); err != nil {
		return err
	}
	return nil
}

func (s *serverStreamImpl[Req, Res]) CloseSend() error {
	return s.call.SendClose(context.Background())
}
