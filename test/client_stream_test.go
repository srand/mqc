package test

import (
	"context"
	"errors"
	"io"
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

type ClientStreamTestSuite struct {
	suite.Suite
	clientConn mqc.Transport
	serverConn mqc.Transport
	serverMock *ClientStreamTestServerMock
	client     ClientStreamTestClient
}

func NewClientStreamTestSuite(clientConn, serverConn mqc.Transport) *ClientStreamTestSuite {
	client := NewClientStreamTestClient(clientConn)

	return &ClientStreamTestSuite{
		clientConn: clientConn,
		serverConn: serverConn,
		serverMock: &ClientStreamTestServerMock{},
		client:     client,
	}
}

// SetupSuite runs once before the suite starts
func (s *ClientStreamTestSuite) SetupSuite() {
	s.T().Log("Setting up ClientStreamTestSuite")
	assert.NotNil(s.T(), s.clientConn)
	assert.NotNil(s.T(), s.serverConn)

	RegisterClientStreamTestServer(s.serverConn, s.serverMock)
	go s.serverConn.Serve()

	time.Sleep(100 * time.Millisecond) // Give the server some time to start
}

// SetupTest runs before each test in the suite
func (s *ClientStreamTestSuite) SetupTest() {
	s.serverMock.ExpectedCalls = nil
	s.serverMock.Calls = nil
}

// TearDown runs after each test in the suite
func (s *ClientStreamTestSuite) TearDown() {
	s.clientConn.Close()

	// Assert that all expectations were met
	s.serverMock.AssertExpectations(s.T())
}

// TearDownSuite runs once after all tests in the suite
func (s *ClientStreamTestSuite) TearDownSuite() {
	s.T().Log("Tearing down ClientStreamTestSuite")
	s.serverConn.Close()
}

func (s *ClientStreamTestSuite) TestStreamSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	requests := []*TestRequest{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}
	expectedReply := &TestReply{Value: 6}

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		stream := args.Get(0).(mqc.ClientStreamServer[TestRequest, TestReply])
		var sum int32
		for {
			req, err := stream.Recv(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(s.T(), err, "Error receiving request")
			sum += req.Value
		}

		expectedReply.Value = sum

		err := stream.SendAndClose(ctx, expectedReply)
		assert.NoError(s.T(), err, "SendAndClose should not return an error")

	}).Return(nil)

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	for _, request := range requests {
		err := stream.Send(ctx, request)
		assert.NoError(s.T(), err)
	}

	reply, err := stream.CloseAndRecv(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), reply)
	assert.Equal(s.T(), expectedReply.Value, reply.Value)
}

func (s *ClientStreamTestSuite) TestStreamServerError_Recv() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	requests := []*TestRequest{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}
	expected := errors.New("server error")

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(expected).Run(func(args mock.Arguments) {
		stream := args.Get(0).(mqc.ClientStreamServer[TestRequest, TestReply])

		for {
			_, err := stream.Recv(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(s.T(), err)
		}
	})

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	for _, request := range requests {
		err := stream.Send(ctx, request)
		assert.NoError(s.T(), err)
	}

	reply, err := stream.CloseAndRecv(ctx)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), reply)
	assert.Equal(s.T(), expected, err)
}

func (s *ClientStreamTestSuite) TestStreamServerError_Send() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 1}
	expected := errors.New("server error")

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(expected)

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	// Wait a moment to ensure the server goroutine is running
	time.Sleep(500 * time.Millisecond)

	err = stream.Send(ctx, request)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), expected, err)
}

func TestClientStreamOverTcp(t *testing.T) {
	// Create two tcp transports
	clientConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewClientStreamTestSuite(clientConn, serverConn))
}

func TestClientStreamOverUnix(t *testing.T) {
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

	suite.Run(t, NewClientStreamTestSuite(clientConn, serverConn))
}

func TestClientStreamOverMqtt(t *testing.T) {
	// Create two mqtt transports
	clientConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewClientStreamTestSuite(clientConn, serverConn))
}

func TestClientStreamOverHttp(t *testing.T) {
	// Create two http transports
	clientConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8081"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8081"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewClientStreamTestSuite(clientConn, serverConn))
}
