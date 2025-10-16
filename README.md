# MQC - An RPC Library

This library provides tools and utilities for implementing Remote Procedure Call (RPC) mechanisms, enabling communication between distributed systems and services.

## Features

- Automatic code generation for client and server stubs using Protocol Buffers
- Customizable transport layers, e.g. MQTT, TCP, Unix sockets, etc.
- Customizable serialization formats, e.g. JSON, Protobuf, etc.
- Support for unary and streaming RPCs
- Error handling and response management

## Installation

```bash
    go get github.com/srand/mqc
```

## Code Generation

To generate the necessary client and server code from your `.proto` files, use the `protoc` command with the `protoc-gen-go-mqc` plugin. Ensure you have the Protocol Buffers compiler (`protoc`) installed.

```bash
    protoc --go-mqc_out=. your_service.proto
```

The `protoc-gen-go-mqc` plugin must be built and available in your `PATH`. You can build it using:

```bash
    go build -o $GOPATH/bin/protoc-gen-go-mqc ./cmd/protoc-gen-go-mqc
```

## Examples

See the [examples](./examples) directory for sample implementations of both client and server using this library.

