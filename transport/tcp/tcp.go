package tcp

import (
	"context"
	"net"
	"time"

	conn_transport "github.com/srand/mqc/transport/conn"

	"github.com/google/uuid"
	"github.com/hashicorp/yamux"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
	"github.com/srand/mqc/transport"
)

type tcpTransport struct {
	id          uuid.UUID
	dialOptions *transport.DialOptions
	conn        net.Conn
	mux         *yamux.Session
	handlers    map[mqc.Method]mqc.ServerHandler
	serializer  serialization.Serializer
}

var _ mqc.Conn = (*tcpTransport)(nil)

func NewTransport(options ...transport.DialOption) (mqc.Conn, error) {
	dialOptions := &transport.DialOptions{
		ConnectTimeout: time.Second * 5,
		CallTimeout:    time.Second * 5,
		Protocol:       "tcp",
	}

	for _, opt := range options {
		opt(dialOptions)
	}

	if len(dialOptions.Addrs) == 0 {
		return nil, mqc.ErrNoAddress
	}

	return &tcpTransport{
		id:          uuid.New(),
		dialOptions: dialOptions,
		handlers:    make(map[mqc.Method]mqc.ServerHandler),
		serializer:  serialization.NewProtoSerializer(),
	}, nil
}

func (t *tcpTransport) ensureConnected() error {
	if t.conn == nil {
		var err error
		t.conn, err = net.Dial(t.dialOptions.Protocol, t.dialOptions.Addrs[0])
		if err != nil {
			return err
		}

		t.mux, err = yamux.Client(t.conn, nil)
		if err != nil {
			t.conn.Close()
			t.conn = nil
			return err
		}

		go t.accept(t.mux)
	}
	return nil
}

func (t *tcpTransport) accept(mux *yamux.Session) error {
	context := context.Background()

	for {
		conn, err := mux.Accept()
		if err != nil {
			return err
		}

		call := conn_transport.NewCall(conn, t.serializer)

		msg, err := call.Recv(context)
		if err != nil {
			conn.Close()
			continue
		}

		if !msg.IsCall() {
			conn.Close()
			continue
		}

		handler, ok := t.handlers[msg.Method()]
		if !ok {
			conn.Close()
			continue
		}

		go func() {
			defer conn.Close()
			handler(call)
		}()
	}
}

func (t *tcpTransport) Close() error {
	if t.mux != nil {
		t.mux.Close()
		t.mux = nil
	}
	if t.conn != nil {
		err := t.conn.Close()
		t.conn = nil
		return err
	}
	return nil
}

func (t *tcpTransport) Invoke(ctx context.Context, method mqc.Method) (mqc.Call, error) {
	if err := t.ensureConnected(); err != nil {
		return nil, err
	}

	conn, err := t.mux.Open()
	if err != nil {
		return nil, err
	}

	call := conn_transport.NewCall(conn, t.serializer)

	err = call.Send(ctx, mqc.NewCallMessage(method))
	if err != nil {
		conn.Close()
		return nil, err
	}

	return call, nil
}

func (t *tcpTransport) RegisterHandler(method mqc.Method, handler mqc.ServerHandler) error {
	t.handlers[method] = handler
	return nil
}

func (t *tcpTransport) UnregisterHandler(method mqc.Method) error {
	delete(t.handlers, method)
	return nil
}

func (t *tcpTransport) Serve() error {
	listener, err := net.Listen(t.dialOptions.Protocol, t.dialOptions.Addrs[0])
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		go func() {
			defer conn.Close()

			// Create a new yamux session for the incoming connection
			session, err := yamux.Server(conn, nil)
			if err != nil {
				conn.Close()
				return
			}
			defer session.Close()

			// Handle incoming streams
			t.accept(session)
		}()
	}
}

func (t *tcpTransport) Serializer() serialization.Serializer {
	return t.serializer
}
