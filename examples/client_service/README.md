Client Service Example
======================

Duplex communication is established between the client and server, allowing for bidirectional data flow. The client implements a service that the server can call back, demonstrating the capability for two-way interaction.

This is useful in scenarios where the server needs to notify the client of events or request data asynchronously. Another use case is NAT traversal, where the client behind a NAT can expose services to the server.

Running the Example
-------------------

To run the example, follow these steps:

1. Start the server:
   ```
   go run examples/client_service/server/main.go
   ```

2. In a separate terminal, start the client:
   ```
   go run examples/client_service/client/main.go
   ```

3. Observe the output in both terminals. Upon establishing the connection, the server will call the client's service and print the message received from the client.

