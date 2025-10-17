WebSocket Example
===================

This example demonstrates the use of HTTP websockets as a transport layer for bidirectional streaming RPCs using MQC. The example implements a simple service that increments integers sent by the client and returns the incremented values back to the client.

Running the Example
-------------------

To run the example, follow these steps:

1. Start the server:
   ```
   go run examples/websocket/server/main.go
   ```

2. In a separate terminal, start the client:
   ```
   go run examples/websocket/client/main.go
   ```

3. Observe the output in both the server and client terminals. The client will send a series of integers to the server, which will increment each integer by 1 and send it back. The client will print the received incremented values.