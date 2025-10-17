package test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/srand/mqc"
	"github.com/srand/mqc/transport"
	"github.com/srand/mqc/transport/http"
	"github.com/srand/mqc/transport/mqtt"
	tpc "github.com/srand/mqc/transport/tcp"
	"github.com/srand/mqc/transport/unix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type RpcTestSuite struct {
	suite.Suite
	clientConn mqc.Transport
	serverConn mqc.Transport
	serverMock *RpcTestServerMock
	client     RpcTestClient
}

func NewRpcTestSuite(clientConn, serverConn mqc.Transport) *RpcTestSuite {
	client := NewRpcTestClient(clientConn)

	return &RpcTestSuite{
		clientConn: clientConn,
		serverConn: serverConn,
		serverMock: &RpcTestServerMock{},
		client:     client,
	}
}

// SetupSuite runs once before the suite starts
func (s *RpcTestSuite) SetupSuite() {
	assert.NotNil(s.T(), s.clientConn)
	assert.NotNil(s.T(), s.serverConn)

	RegisterRpcTestServer(s.serverConn, s.serverMock)
	go func() {
		assert.NoError(s.T(), s.serverConn.Serve())
	}()
	time.Sleep(100 * time.Millisecond) // Give the server some time to start
}

// SetupTest runs before each test in the suite
func (s *RpcTestSuite) SetupTest() {
	s.serverMock.ExpectedCalls = nil
	s.serverMock.Calls = nil
}

// TearDown runs after each test in the suite
func (s *RpcTestSuite) TearDown() {
	s.clientConn.Close()

	// Assert that all expectations were met
	s.serverMock.AssertExpectations(s.T())
}

// TearDownSuite runs once after all tests in the suite
func (s *RpcTestSuite) TearDownSuite() {
	s.serverConn.Close()
}

func (s *RpcTestSuite) TestRpc() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		assert.Equal(s.T(), request.Value, args.Get(0).(*TestRequest).Value)
	}).Return(expected, nil)

	resp, err := s.client.Rpc(ctx, request)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), resp)
	assert.Equal(s.T(), expected.Value, resp.Value)
}

func (s *RpcTestSuite) TestRpcServerError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 42}
	expected := errors.New("server error")

	// Setup expected call
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Return(nil, expected)

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
	assert.Equal(s.T(), expected, err)
}

func (s *RpcTestSuite) TestRpcClientTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	// Setup expected call with delay
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(500 * time.Millisecond) // Simulate processing delay
	}).Return(expected, nil)

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
	assert.True(s.T(), errors.Is(err, context.DeadlineExceeded))
}

func (s *RpcTestSuite) TestRpcClientCancel() {
	ctx, cancel := context.WithCancel(context.Background())

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	// Setup expected call with delay
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(200 * time.Millisecond) // Simulate processing delay
	}).Return(expected, nil)

	// Cancel the context immediately
	cancel()

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
	assert.True(s.T(), errors.Is(err, context.Canceled))
}

func (s *RpcTestSuite) TestRpcClientCancelAfterSend() {
	ctx, cancel := context.WithCancel(context.Background())

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	// Setup expected call with delay
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(200 * time.Millisecond) // Simulate processing delay
	}).Return(expected, nil)

	// Start a goroutine to cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
	assert.True(s.T(), errors.Is(err, context.Canceled))
}

func (s *RpcTestSuite) TestRpcNilRequest() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// No need to setup expected call since the request is nil

	stream, err := s.client.Rpc(ctx, nil)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
}

func (s *RpcTestSuite) TestRpcClientDisconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	// Close the client connection to simulate disconnect
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		s.clientConn.Close()
	}).Return(expected, nil)

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
}

// Currently disabled, because the server does not handle panics gracefully
func (s *RpcTestSuite) TestRpcServerDisconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 42}
	expected := &TestReply{Value: 43}

	// Close the server connection to simulate disconnect
	s.serverMock.On("Rpc", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		panic("simulated server panic")
	}).Return(expected, nil)

	stream, err := s.client.Rpc(ctx, request)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), stream)
}

func TestSimpleRpcOverTcp(t *testing.T) {
	// Create two tcp transports
	clientConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewRpcTestSuite(clientConn, serverConn))
}
func TestSimpleRpcOverUnix(t *testing.T) {
	// Remove socket if it exists
	os.Remove("/tmp/mqc.sock")

	// Create two unix transports
	clientConn, err := unix.NewTransport(transport.WithAddress("/tmp/mqc.sock"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := unix.NewTransport(transport.WithAddress("/tmp/mqc.sock"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewRpcTestSuite(clientConn, serverConn))
}

func TestSimpleRpcOverMqtt(t *testing.T) {
	// Create two mqtt transports
	clientConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewRpcTestSuite(clientConn, serverConn))
}

func TestSimpleRpcOverHttp(t *testing.T) {
	// Create two http transports
	clientConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8083"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8083"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewRpcTestSuite(clientConn, serverConn))
}
