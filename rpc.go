package mqc

import (
	"context"

	"github.com/srand/mqc/serialization"
)

// Rpc performs a remote procedure call to the specified method with the given request.
// It sends the request and waits for a response, returning the response object or an error.
func Rpc[Req any, Res any](ctx context.Context, transport Transport, method *Method, req *Req) (*Res, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if req == nil {
		return nil, ErrNilRequest
	}

	serializer := transport.Serializer()

	// Create a new connection for the RPC call
	stream, err := transport.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// Marshal request
	data, err := serializer.Marshal(req)
	if err != nil {
		return nil, err
	}

	// Send request
	if err := stream.Send(ctx, data); err != nil {
		return nil, err
	}

	// Receive response
	data, err = stream.Recv(ctx)
	if err != nil {
		return nil, err
	}

	var res Res

	err = serializer.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

// RpcServer handles an incoming RPC call on the server side.
// It receives the request, processes it using the provided handler function,
// and sends back the response or an error.
func RpcServer[Req any, Res any](conn Conn, serializer serialization.Serializer, handler func(req *Req) (*Res, error)) error {
	ctx := context.Background()

	// Receive request
	data, err := conn.Recv(ctx)
	if err != nil {
		return err
	}

	// Unmarshal request
	var req Req
	err = serializer.Unmarshal(data, &req)
	if err != nil {
		return err
	}

	// Handle request
	res, err := handler(&req)
	if err != nil {
		return err
	}

	data, err = serializer.Marshal(res)
	if err != nil {
		return err
	}

	// Send response
	return conn.Send(ctx, data)
}
