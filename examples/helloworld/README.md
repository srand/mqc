Hello World Example
===================

This example demonstrates a simple "Hello, World!" service using TCP as the transport layer.

To run the example, start the server in one terminal:

```
go run server/main.go
```

Then, in another terminal, run the client:

```
go run client/main.go
```

The client will connect to the server, send a greeting request, and print the server's response.
