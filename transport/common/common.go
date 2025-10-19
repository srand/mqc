package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/yamux"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
	"github.com/srand/mqc/transport"
)

type BaseTransport struct {
	Handlers  map[mqc.Method]mqc.MethodHandler
	Options   transport.TransportOptions
	Serialize serialization.Serializer
}

func (t *BaseTransport) RegisterHandler(method *mqc.Method, handler mqc.MethodHandler) error {
	t.Handlers[*method] = handler
	return nil
}

func (t *BaseTransport) UnregisterHandler(method *mqc.Method) error {
	delete(t.Handlers, *method)
	return nil
}

func (t *BaseTransport) Serializer() serialization.Serializer {
	return t.Serialize
}

func (t *BaseTransport) AcceptMux(mux *yamux.Session) error {
	ctx := context.Background()

	for {
		conn, err := mux.Accept()
		if err != nil {
			return err
		}

		go func() {
			call := NewConn(conn, t.Serialize)

			method, err := call.RecvMethod(ctx)
			if err != nil {
				conn.Close()
				return
			}

			handler, ok := t.Handlers[*method]
			if !ok {
				conn.Close()
				return
			}

			// Ensure that any panic is recovered
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered in server goroutine:", r)
				}
			}()

			defer conn.Close()

			err = handler(call)
			if err != nil {
				call.SendError(ctx, err)
			}
		}()
	}
}

func (t *BaseTransport) InvokeMux(ctx context.Context, mux *yamux.Session, method *mqc.Method) (mqc.Conn, error) {
	conn, err := mux.Open()
	if err != nil {
		return nil, err
	}

	call := NewConn(conn, t.Serialize)

	err = call.SendMethod(ctx, method)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return call, nil
}
