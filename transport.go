package mqc

import (
	"context"

	"github.com/srand/mqc/serialization"
)

type MethodHandler func(stream Conn) error

// Transport is a communication transport for RPC calls.
type Transport interface {
	// Close closes the transport.
	// Any blocked Read or Write operations will be unblocked and return errors.
	Close() error

	// Dial establishes a connection to the remote server.
	// If the transport is already connected, it returns an error.
	// It is not necessary to call Dial before Invoke, as Invoke will dial automatically.
	Dial() error

	// Invoke creates a new connection object for the given method.
	Invoke(ctx context.Context, method *Method) (Conn, error)

	// RegisterHandler registers a new handler for the given method.
	RegisterHandler(method *Method, handler MethodHandler) error

	// Serializer returns the serializer used for marshaling and unmarshaling messages.
	Serializer() serialization.Serializer

	// Serve starts the server to accept incoming connections and handle requests.
	Serve() error
}
