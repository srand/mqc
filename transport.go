package mqc

import (
	"context"

	"github.com/srand/mqc/serialization"
)

type MethodHandler func(stream Conn) error

// Transport is a communication transport for RPC calls.
type Transport interface {
	// RegisterHandler registers a new handler for the given method.
	RegisterHandler(method Method, handler MethodHandler) error

	// Invoke creates a new connection object for the given method.
	Invoke(ctx context.Context, method Method) (Conn, error)

	// Close closes the transport.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// Serve starts the server to accept incoming connections and handle requests.
	Serve() error

	// Serializer returns the serializer used for marshaling and unmarshaling messages.
	Serializer() serialization.Serializer
}
