package http

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/yamux"
	"github.com/srand/mqc"
	"golang.org/x/net/websocket"
)

type httpHandler struct {
	handler   websocket.Handler
	transport mqc.Transport
}

var _ http.Handler = (*httpHandler)(nil)

func NewHandler(transport mqc.Transport) http.Handler {
	t, ok := transport.(*websocketTransport)
	if !ok {
		panic("transport must be a websocket transport")
	}

	handler := websocket.Handler(func(ws *websocket.Conn) {
		mux, err := yamux.Server(ws, nil)
		if err != nil {
			ws.Close()
			return
		}
		defer mux.Close()

		if err := t.AcceptMux(mux); err != nil {
			fmt.Println("Error accepting mux:", err)
			return
		}
	})
	return &httpHandler{
		handler:   handler,
		transport: transport,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
