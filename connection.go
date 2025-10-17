package mqc

import (
	"context"
)

type Conn interface {
	// Send sends data to the stream.
	Send(ctx context.Context, data []byte) error

	// SendClose sends a close message to the stream.
	SendClose(ctx context.Context) error

	// Recv receives data from the stream.
	Recv(ctx context.Context) ([]byte, error)

	// Close closes the stream.
	Close() error
}
