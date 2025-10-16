//go:generate protoc -I.. --go_out=../../.. ../entropy.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../entropy.proto

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"loadbalancer"
	"log"
	"math/rand"
	"time"

	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/mqtt"
)

var (
	addr = flag.String("addr", "localhost:1883", "the address to connect to")
)

func init() {
	flag.Parse()
}

func main() {
	// Create a new transport.
	// The address should point to a MQTT broker.
	transport, err := mqtt.NewTransport(
		transport.WithAddress(*addr),
	)
	if err != nil {
		log.Fatalf("could not create transport: %v", err)
	}
	defer transport.Close()

	// Create a new service client
	client := loadbalancer.NewEntropyClient(transport)
	ctx := context.Background()

	// Initial timeout for the request if no server is available
	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Request between 5 and 15 random numbers
	request := &loadbalancer.NumberRequest{Count: int32(5 + rand.Intn(10))}

	// Initiate the request
	stream, err := client.GenerateIntegers(timeout, request)
	if err != nil {
		log.Fatalf("could not generate integers: %v", err)
	}

	for {
		// Timeout for each number received
		timeout, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		// Wait for a number
		reply, err := stream.Recv(timeout)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Fatalf("could not receive: %v", err)
		}
		fmt.Printf("Number: %d\n", reply.Value)
	}
}
