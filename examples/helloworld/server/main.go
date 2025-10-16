//go:generate protoc -I.. --go_out=../../.. ../helloworld.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../helloworld.proto

package main

import (
	"flag"
	"helloworld"

	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/tcp"
)

var (
	addr = flag.String("addr", "localhost:12345", "the address to connect to")
)

type server struct {
	helloworld.UnimplementedGreeterServer
}

func (s *server) SayHello(req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hello " + req.Name}, nil
}

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

	helloworld.RegisterGreeterServer(conn, &server{})
	conn.Serve()
}
