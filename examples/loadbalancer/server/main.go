//go:generate protoc -I.. --go_out=../../.. ../entropy.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../entropy.proto

package main

import (
	"context"
	"flag"
	"loadbalancer"
	"math/rand"
	"time"

	"github.com/srand/mqc"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/mqtt"
)

var (
	addr = flag.String("addr", "localhost:1883", "the address to connect to")
)

func init() {
	flag.Parse()
}

type server struct {
	loadbalancer.UnimplementedEntropyServer
}

// GenerateIntegers generates a stream of random integers.
// If multiple servers are running, the MQTT broker will load balance
// the requests between them.
func (s *server) GenerateIntegers(req *loadbalancer.NumberRequest, stream mqc.ServerStreamServer[loadbalancer.NumberReply]) error {
	ctx := context.Background()

	println("Generating ", req.Count, " random numbers")

	for i := 0; i < int(req.Count); i++ {
		// Create a reply with a random number
		reply := &loadbalancer.NumberReply{Value: rand.Int31()}

		// Send the reply to the client
		err := stream.Send(ctx, reply)
		if err != nil {
			return err
		}

		// Simulate some processing time
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}

	return nil
}

func main() {
	// Create a new transport.
	// The address should point to a MQTT broker.
	conn, err := mqtt.NewTransport(
		transport.WithAddress(*addr),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Register the entropy service on the server
	loadbalancer.RegisterEntropyServer(conn, &server{})

	// Start the server
	panic(conn.Serve().Error())
}
