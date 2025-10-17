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

type ServerStreamTestSuite struct {
	suite.Suite
	clientConn mqc.Transport
	serverConn mqc.Transport
	serverMock *ServerStreamTestServerMock
	client     ServerStreamTestClient
}

func NewServerStreamTestSuite(clientConn, serverConn mqc.Transport) *ServerStreamTestSuite {
	client := NewServerStreamTestClient(clientConn)

	return &ServerStreamTestSuite{
		clientConn: clientConn,
		serverConn: serverConn,
		serverMock: &ServerStreamTestServerMock{},
		client:     client,
	}
}

// SetupSuite runs once before the suite starts
func (s *ServerStreamTestSuite) SetupSuite() {
	s.T().Log("Setting up ServerStreamTestSuite")
	assert.NotNil(s.T(), s.clientConn)
	assert.NotNil(s.T(), s.serverConn)

	RegisterServerStreamTestServer(s.serverConn, s.serverMock)
	go s.serverConn.Serve()

	time.Sleep(100 * time.Millisecond) // Give the server some time to start
}

// SetupTest runs before each test in the suite
func (s *ServerStreamTestSuite) SetupTest() {
	s.serverMock.ExpectedCalls = nil
	s.serverMock.Calls = nil
}

// TearDown runs after each test in the suite
func (s *ServerStreamTestSuite) TearDown() {
	s.clientConn.Close()

	// Assert that all expectations were met
	s.serverMock.AssertExpectations(s.T())
}

// TearDownSuite runs once after all tests in the suite
func (s *ServerStreamTestSuite) TearDownSuite() {
	s.T().Log("Tearing down ServerStreamTestSuite")
	s.serverConn.Close()
}

func (s *ServerStreamTestSuite) TestStreamSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	request := &TestRequest{Value: 0}
	expectedReplies := []*TestReply{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		request := args.Get(0).(*TestRequest)
		assert.Equal(s.T(), int32(0), request.Value, "Request value should match")

		stream := args.Get(1).(mqc.ServerStreamServer[TestReply])
		for _, reply := range expectedReplies {
			if err := stream.Send(ctx, reply); err != nil {
				s.T().Errorf("Failed to send reply: %v", err)
			}
		}
	}).Return(nil)

	stream, err := s.client.Stream(ctx, request)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	var replies []*TestReply
	for {
		reply, err := stream.Recv(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(s.T(), err, "Error receiving reply")
		replies = append(replies, reply)
	}

	assert.Equal(s.T(), len(expectedReplies), len(replies), "Number of replies should match")
	for i, reply := range replies {
		assert.Equal(s.T(), expectedReplies[i].Value, reply.Value, "Reply value should match")
	}
}

func (s *ServerStreamTestSuite) TestStreamServerError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request := &TestRequest{Value: 0}
	err := errors.New("server error")

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(err)

	stream, err := s.client.Stream(ctx, request)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	reply, err := stream.Recv(ctx)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), reply)
	assert.Equal(s.T(), "server error", err.Error())
}

func TestServerStreamOverTcp(t *testing.T) {
	// Create two tcp transports
	clientConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewServerStreamTestSuite(clientConn, serverConn))
}

func TestServerStreamOverUnix(t *testing.T) {
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

	suite.Run(t, NewServerStreamTestSuite(clientConn, serverConn))
}

func TestServerStreamOverMqtt(t *testing.T) {
	// Create two mqtt transports
	clientConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewServerStreamTestSuite(clientConn, serverConn))
}

func TestServerStreamOverHttp(t *testing.T) {
	// Create two http transports
	clientConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8082"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := http.NewWebSocketTransport(
		transport.WithAddress("ws://localhost:8082"),
		transport.WithOrigin("http://localhost/"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewServerStreamTestSuite(clientConn, serverConn))
}
