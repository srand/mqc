package mqc

import (
	"context"

	"github.com/srand/mqc/serialization"
)

type Call interface {
	// Send sends a message to the stream.
	Send(ctx context.Context, msg *Message) error

	// Recv receives a message from the stream.
	Recv(ctx context.Context) (*Message, error)

	// Close closes the stream.
	Close() error
}

type ServerHandler func(stream Call) error

// Conn is a generic stream-oriented network connection.
//
// Multiple goroutines may invoke methods on a Conn simultaneously.
type Conn interface {
	// RegisterHandler registers a new handler for the given method.
	RegisterHandler(method Method, handler ServerHandler) error

	// UnregisterHandler unregisters the handler for the given method.
	UnregisterHandler(method Method) error

	// Invoke creates a new call object for the given method.
	Invoke(ctx context.Context, method Method) (Call, error)

	// Close closes the connection.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// Serve starts the server to accept incoming connections and handle requests.
	Serve() error

	// Serializer returns the serializer used for marshaling and unmarshaling messages.
	Serializer() serialization.Serializer
}
