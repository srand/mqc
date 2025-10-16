Load Balancer Example
=====================

This example demonstrates a simple load balanced service using MQTT as the transport layer.

Multiple server instances can be started, each registering the same service. The MQTT broker will then distribute client requests among the available servers.

To run the example, start an MQTT broker (e.g., Mosquitto) and then run multiple instances of the server:

```
go run server/main.go
```

Then, in another terminal, run the client:

```
go run client/main.go
```

By default, these examples connect to a local MQTT broker at `tcp://localhost:1883`. You can change the broker address by setting the `-addr` flag when starting the server and client.
