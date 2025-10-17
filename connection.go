package mqc

import (
	"context"
)

type Conn interface {
	// Send sends a message to the stream.
	Send(ctx context.Context, msg *Message) error

	// Recv receives a message from the stream.
	Recv(ctx context.Context) (*Message, error)

	// Close closes the stream.
	Close() error
}
