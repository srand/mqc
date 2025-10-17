//go:generate protoc -I.. --go_out=../../.. ../bidi.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../bidi.proto
package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"websocket"

	"github.com/srand/mqc"
	mqc_http "github.com/srand/mqc/transport/http"
)

type incrementerServer struct {
	websocket.UnimplementedIncrementerServer
}

func (s *incrementerServer) Increment(stream mqc.BidiStreamServer[websocket.Integer, websocket.Integer]) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		req, err := stream.Recv(ctx)
		if err != nil {
			return err
		}

		resp := &websocket.Integer{Value: req.Value + 1}
		if err := stream.Send(ctx, resp); err != nil {
			return err
		}
	}
}

func main() {
	transport, err := mqc_http.NewWebSocketTransport()
	if err != nil {
		log.Fatalf("failed to create websocket transport: %v", err)
	}

	websocket.RegisterIncrementerServer(transport, &incrementerServer{})

	http.Handle("/ws", mqc_http.NewHandler(transport))
	http.ListenAndServe(":8080", nil)
}
