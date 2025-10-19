PubSub Example
=================

This example demonstrates how to use the `mqc` library to implement a simple publish-subscribe messaging system. MQTT is used as the transport layer. If multiple consumers are running, they will each receive messages published by a publisher.

The protobuf compiler plugin treats services with bi-directional streaming methods and identical input and output message types as PubSub services. For such methods, the generated code will provide specialized client and server implementations for publishing and consuming messages.

To run the example, start an MQTT broker (e.g., Mosquitto) and then run multiple instances of the consumer:

```
go run consumer/main.go
```

Then, in another terminal, run the publisher:

```
go run publisher/main.go
```

By default, these examples connect to a local MQTT broker at `tcp://localhost:1883`. You can change the broker address by setting the `-addr` flag when starting the server and client.

