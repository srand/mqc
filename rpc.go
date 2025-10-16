package mqc

import (
	"context"

	"github.com/srand/mqc/serialization"
)

// Rpc performs a remote procedure call to the specified method with the given request.
// It sends the request and waits for a response, returning the response object or an error.
func Rpc[Req any, Res any](ctx context.Context, conn Conn, method Method, req *Req) (*Res, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if req == nil {
		return nil, ErrNilRequest
	}

	serializer := conn.Serializer()

	stream, err := conn.Invoke(ctx, method)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	// Marshal request
	msg, err := NewDataMessage(req, serializer)
	if err != nil {
		return nil, err
	}

	// Send request
	if err := stream.Send(ctx, msg); err != nil {
		return nil, err
	}

	// Receive response
	res, err := stream.Recv(ctx)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, res.Error()
	}

	resObj, err := GetMessageData[Res](res, serializer)
	if err != nil {
		return nil, err
	}
	return resObj, nil
}

// RpcServer handles an incoming RPC call on the server side.
// It receives the request, processes it using the provided handler function,
// and sends back the response or an error.
func RpcServer[Req any, Res any](call Call, serializer serialization.Serializer, handler func(req *Req) (*Res, error)) error {
	ctx := context.Background()

	// Receive request
	req, err := call.Recv(ctx)
	if err != nil {
		return err
	}

	reqObj, err := GetMessageData[Req](req, serializer)
	if err != nil {
		return err
	}

	// Handle request
	resObj, err := handler(reqObj)
	if err != nil {
		return err
	}

	// Marshal response
	resMsg, err := NewDataMessage(resObj, serializer)
	if err != nil {
		return err
	}

	// Send response
	return call.Send(ctx, resMsg)
}
