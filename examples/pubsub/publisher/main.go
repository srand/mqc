//go:generate protoc -I.. --go_out=../../.. ../weather.proto
//go:generate protoc -I.. --go-mqc_out=../../.. ../weather.proto

package main

import (
	"context"
	"time"

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
	weather := pubsub.NewWeatherPublisher(transport)
	stream, err := weather.Update(ctx)
	if err != nil {
		panic(err)
	}

	for {
		update := &pubsub.WeatherUpdate{
			City:        "San Francisco",
			Temperature: 20,
			Condition:   "Sunny",
		}

		if err := stream.Send(ctx, update); err != nil {
			panic(err)
		}

		time.Sleep(2 * time.Second)
	}
}
