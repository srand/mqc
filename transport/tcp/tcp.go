package tcp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/common"
)

type tcpTransport struct {
	common.BaseTransport

	conn net.Conn
	mux  *yamux.Session
}

var _ mqc.Transport = (*tcpTransport)(nil)

func NewTransport(options ...transport.TransportOption) (mqc.Transport, error) {
	transportOptions := &transport.TransportOptions{
		ConnectTimeout: time.Second * 5,
		CallTimeout:    time.Second * 5,
		Protocol:       "tcp",
	}

	for _, opt := range options {
		if err := opt(transportOptions); err != nil {
			return nil, err
		}
	}

	if len(transportOptions.Addrs) == 0 {
		return nil, mqc.ErrNoAddress
	}

	return &tcpTransport{
		BaseTransport: common.BaseTransport{
			Options:   *transportOptions,
			Handlers:  make(map[mqc.Method]mqc.MethodHandler),
			Serialize: serialization.NewProtoSerializer(),
		},
	}, nil
}

func (t *tcpTransport) ensureConnected() error {
	if t.conn == nil {
		var err error

		if t.Options.TlsConfig != nil {
			t.conn, err = tls.Dial(t.Options.Protocol, t.Options.Addrs[0], t.Options.TlsConfig)
		} else {
			t.conn, err = net.Dial(t.Options.Protocol, t.Options.Addrs[0])
		}
		if err != nil {
			return err
		}

		t.mux, err = yamux.Client(t.conn, nil)
		if err != nil {
			t.conn.Close()
			t.conn = nil
			return err
		}

		go t.AcceptMux(t.mux)
	}
	return nil
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

func (t *tcpTransport) Dial() error {
	if t.conn != nil {
		return fmt.Errorf("transport is already connected")
	}
	return t.ensureConnected()
}

func (t *tcpTransport) Invoke(ctx context.Context, method mqc.Method) (mqc.Conn, error) {
	if err := t.ensureConnected(); err != nil {
		return nil, err
	}
	return t.InvokeMux(ctx, t.mux, method)
}

func (t *tcpTransport) Serve() error {
	if t.conn != nil {
		return fmt.Errorf("transport is already connected")
	}

	var err error
	var listener net.Listener

	if t.Options.TlsConfig != nil {
		listener, err = tls.Listen(t.Options.Protocol, t.Options.Addrs[0], t.Options.TlsConfig)
	} else {
		listener, err = net.Listen(t.Options.Protocol, t.Options.Addrs[0])
	}
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

			clientTransport := &tcpTransport{
				BaseTransport: t.BaseTransport,
				conn:          conn,
				mux:           session,
			}

			if t.Options.OnConnect != nil {
				t.Options.OnConnect(clientTransport)
			}

			// Handle incoming streams
			t.AcceptMux(session)
		}()
	}
}

func (t *tcpTransport) Serializer() serialization.Serializer {
	return t.Serialize
}
