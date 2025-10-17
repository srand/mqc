//go:generate protoc -I.. --go_out=../../.. ../bidi.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../bidi.proto

package main

import (
	"context"
	"log"
	"time"
	"websocket"

	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/http"
)

func main() {
	transport, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8080/ws"),
		transport.WithOrigin("http://localhost:8080"),
	)
	if err != nil {
		log.Fatalf("failed to create websocket transport: %v", err)
	}

	client := websocket.NewIncrementerClient(transport)
	stream, err := client.Increment(context.Background())
	if err != nil {
		log.Fatalf("failed to start increment stream: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := range []int{1, 4, 5, 3, 4, 5} {
		if err := stream.Send(ctx, &websocket.Integer{Value: int32(i)}); err != nil {
			log.Fatalf("failed to send integer: %v", err)
		}

		resp, err := stream.Recv(ctx)
		if err != nil {
			log.Fatalf("failed to receive incremented integer: %v", err)
		}

		log.Printf("Sent: %d, Received: %d", i, resp.Value)
	}

	if err := stream.CloseSend(); err != nil {
		log.Fatalf("failed to close stream: %v", err)
	}
}
