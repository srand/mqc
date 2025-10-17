//go:generate protoc -I.. --go_out=../../.. --go-mqc_out=../../.. ../service.proto

package main

import (
	"client_service"
	"context"
	"flag"

	"github.com/srand/mqc"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/tcp"
)

var (
	addr = flag.String("addr", "localhost:12345", "the address to connect to")
)

func init() {
	flag.Parse()
}

func main() {
	transport, err := tcp.NewTransport(transport.WithAddress(*addr))
	if err != nil {
		panic(err)
	}

	transport.OnConnect(func(transport mqc.Transport) {
		println("A client connected to the server")

		ctx := context.Background()
		client := client_service.NewEchoClient(transport)

		resp, err := client.Echo(ctx, &client_service.EchoRequest{Message: "Hello from client!"})
		if err != nil {
			println("Error calling Echo:", err.Error())
			return
		}
		println("Received response from service:", resp.Message)

	})

	panic(transport.Serve())
}
