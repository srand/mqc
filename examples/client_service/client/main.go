//go:generate protoc -I.. --go_out=../../.. --go-mqc_out=../../.. ../service.proto

package main

import (
	"client_service"
	"flag"
	"time"

	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/tcp"
)

var (
	addr = flag.String("addr", "localhost:12345", "the address to connect to")
	done = make(chan struct{})
)

func init() {
	flag.Parse()
}

type service struct {
	client_service.UnimplementedEchoServer
}

func (s *service) Echo(req *client_service.EchoRequest) (*client_service.EchoReply, error) {
	println("Server called client service with message:", req.Message)
	close(done)
	return &client_service.EchoReply{Message: req.Message}, nil
}

func main() {
	transport, err := tcp.NewTransport(transport.WithAddress(*addr))
	if err != nil {
		panic(err)
	}

	client_service.RegisterEchoServer(transport, &service{})

	if err := transport.Dial(); err != nil {
		panic(err)
	}

	<-done
	time.Sleep(time.Second)
}
