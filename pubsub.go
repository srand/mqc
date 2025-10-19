package mqc

import (
	"context"
	"errors"
	"io"

	"github.com/srand/mqc/serialization"
)

type PubSubClient[T any] interface {
	Recv(ctx context.Context) (*T, error)
	Send(ctx context.Context, req *T) error
}

type pubsubImpl[T any] struct {
	call       Conn
	serializer serialization.Serializer
	eof        bool
}

func NewPubSubClient[T any](ctx context.Context, transport Transport, method *Method) (PubSubClient[T], error) {
	call, err := transport.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}

	return &pubsubImpl[T]{call: call, serializer: transport.Serializer()}, nil
}

func (s *pubsubImpl[T]) Recv(ctx context.Context) (*T, error) {
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

	return unmarshal[T](s.serializer, data)
}

func (s *pubsubImpl[T]) Send(ctx context.Context, req *T) error {
	data, err := marshal[T](s.serializer, req)
	if err != nil {
		return err
	}
	if err := s.call.Send(ctx, data); err != nil {
		return err
	}
	return nil
}
