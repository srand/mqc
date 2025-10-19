//go:generate protoc -I.. --go_out=../../.. ../weather.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../weather.proto

package main

import (
	"context"
	"fmt"

	"github.com/srand/mqc/examples/pubsub"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/mqtt"
)

func main() {
	transport, err := mqtt.NewTransport(
		transport.WithAddress("tcp://localhost:1883"),
	)
	if err != nil {
		panic(err)
	}
	defer transport.Close()

	ctx := context.Background()
	weather := pubsub.NewWeatherConsumer(transport)
	stream, err := weather.Update(ctx)
	if err != nil {
		panic(err)
	}

	for {
		update, err := stream.Recv(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Received weather update: %s, %.1fC, %s\n", update.City, update.Temperature, update.Condition)
	}
}
