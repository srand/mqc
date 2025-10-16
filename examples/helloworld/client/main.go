//go:generate protoc -I.. --go_out=../../.. ../helloworld.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../helloworld.proto

package main

import (
	"context"
	"flag"
	"helloworld"

	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/tcp"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:12345", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func init() {
	flag.Parse()
}

func main() {
	conn, err := tcp.NewTransport(
		transport.WithAddress(*addr),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)
	ctx := context.Background()

	resp, err := client.SayHello(ctx, &helloworld.HelloRequest{Name: *name})
	if err != nil {
		panic(err)
	}
	println("Greeting:", resp.Message)
}
