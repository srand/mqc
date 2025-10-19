package http

import (
	"context"
	"net"
	"net/http"
	"net/url"

	"github.com/hashicorp/yamux"
	"github.com/srand/mqc"
	"github.com/srand/mqc/serialization"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/common"
	"golang.org/x/net/websocket"
)

// A websocket transport
type websocketTransport struct {
	common.BaseTransport
	conn net.Conn
	mux  *yamux.Session
}

var _ mqc.Transport = (*websocketTransport)(nil)

// NewWebSocketTransport creates a new WebSocket transport with the given options.
func NewWebSocketTransport(options ...transport.TransportOption) (mqc.Transport, error) {
	transportOptions := &transport.TransportOptions{}

	for _, opt := range options {
		if err := opt(transportOptions); err != nil {
			return nil, err
		}
	}

	return &websocketTransport{
		BaseTransport: common.BaseTransport{
			Handlers:  make(map[mqc.Method]mqc.MethodHandler),
			Options:   *transportOptions,
			Serialize: serialization.NewJSONSerializer(),
		},
	}, nil
}

func (t *websocketTransport) ensureConnected() error {
	ws, err := websocket.Dial(t.Options.Addrs[0], "", t.Options.Origin)
	if err != nil {
		return err
	}
	t.conn = ws

	// Create yamux session
	mux, err := yamux.Client(ws, nil)
	if err != nil {
		ws.Close()
		return err
	}
	t.mux = mux

	return nil
}

// Close closes the transport and releases any resources.
func (t *websocketTransport) Close() error {
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

func (t *websocketTransport) Dial() error {
	return t.ensureConnected()
}

// Serve starts the server to accept incoming connections and handle requests.
func (t *websocketTransport) Serve() error {
	url, err := url.ParseRequestURI(t.Options.Addrs[0])
	if err != nil {
		return err
	}

	path := url.Path
	if path == "" {
		path = "/"
	}

	mux := http.NewServeMux()
	mux.Handle(path, NewHandler(t))

	server := http.Server{
		Addr:      url.Host,
		Handler:   mux,
		TLSConfig: t.Options.TlsConfig,
	}

	return server.ListenAndServe()
}

func (t *websocketTransport) Invoke(ctx context.Context, method *mqc.Method) (mqc.Conn, error) {
	if method.IsPubSub() {
		return nil, mqc.ErrPubSubNotSupported
	}

	if err := t.ensureConnected(); err != nil {
		return nil, err
	}

	return t.InvokeMux(ctx, t.mux, method)
}
