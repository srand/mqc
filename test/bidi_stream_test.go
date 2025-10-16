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
	"github.com/srand/mqc/transport/mqtt"
	tpc "github.com/srand/mqc/transport/tcp"
	"github.com/srand/mqc/transport/unix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BidiStreamTestSuite struct {
	suite.Suite
	clientConn mqc.Conn
	serverConn mqc.Conn
	serverMock *BidiStreamTestServerMock
	client     BidiStreamTestClient
}

func NewBidiStreamTestSuite(clientConn, serverConn mqc.Conn) *BidiStreamTestSuite {
	client := NewBidiStreamTestClient(clientConn)

	return &BidiStreamTestSuite{
		clientConn: clientConn,
		serverConn: serverConn,
		serverMock: &BidiStreamTestServerMock{},
		client:     client,
	}
}

// SetupSuite runs once before the suite starts
func (s *BidiStreamTestSuite) SetupSuite() {
	s.T().Log("Setting up BidiStreamTestSuite")
	assert.NotNil(s.T(), s.clientConn)
	assert.NotNil(s.T(), s.serverConn)

	RegisterBidiStreamTestServer(s.serverConn, s.serverMock)
	go s.serverConn.Serve()

	time.Sleep(100 * time.Millisecond) // Give the server some time to start
}

// SetupTest runs before each test in the suite
func (s *BidiStreamTestSuite) SetupTest() {
	s.serverMock.ExpectedCalls = nil
	s.serverMock.Calls = nil
}

// TearDown runs after each test in the suite
func (s *BidiStreamTestSuite) TearDown() {
	s.clientConn.Close()

	// Assert that all expectations were met
	s.serverMock.AssertExpectations(s.T())
}

// TearDownSuite runs once after all tests in the suite
func (s *BidiStreamTestSuite) TearDownSuite() {
	s.T().Log("Tearing down BidiStreamTestSuite")
	s.serverConn.Close()
}

func (s *BidiStreamTestSuite) TestStreamSuccess() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	requests := []*TestRequest{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}
	replies := []*TestReply{
		{Value: 10},
		{Value: 20},
		{Value: 30},
	}

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		stream := args.Get(0).(mqc.BidiStreamServer[TestRequest, TestReply])

		for {
			req, err := stream.Recv(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(s.T(), err)

			var reply *TestReply
			switch req.Value {
			case 1:
				reply = replies[0]
			case 2:
				reply = replies[1]
			case 3:
				reply = replies[2]
			default:
				s.T().Fatalf("Unexpected request value: %d", req.Value)
			}

			err = stream.Send(ctx, reply)
			assert.NoError(s.T(), err)
		}

		_, err := stream.Recv(ctx)
		assert.Equal(s.T(), io.EOF, err, "Recv after CloseSend should return EOF")
	})

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	for _, request := range requests {
		err := stream.Send(ctx, request)
		assert.NoError(s.T(), err)
	}

	for expected := range replies {
		reply, err := stream.Recv(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(s.T(), err)
		assert.NotNil(s.T(), reply)
		assert.Equal(s.T(), replies[expected].Value, reply.Value)
	}

	err = stream.CloseSend()
	assert.NoError(s.T(), err)

	time.Sleep(100 * time.Millisecond) // Give some time for the server to process the close

	// Read after close should return EOF
	_, err = stream.Recv(ctx)
	assert.Equal(s.T(), io.EOF, err, "Recv after CloseSend should return EOF")
}

func (s *BidiStreamTestSuite) TestStreamSuccessServerClose() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	requests := []*TestReply{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}
	replies := []*TestRequest{
		{Value: 10},
		{Value: 20},
		{Value: 30},
	}

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		stream := args.Get(0).(mqc.BidiStreamServer[TestRequest, TestReply])

		// Server sends requests, closes and waits for replies
		for i := range requests {
			err := stream.Send(ctx, requests[i])
			assert.NoError(s.T(), err)
		}

		err := stream.CloseSend()
		assert.NoError(s.T(), err)

		for i := range replies {
			reply, err := stream.Recv(ctx)
			if errors.Is(err, io.EOF) {
				break
			}
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), replies[i].Value, reply.Value)
		}
	})

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	for i := range requests {
		resp, err := stream.Recv(ctx)
		assert.NoError(s.T(), err)
		assert.Equal(s.T(), requests[i].Value, resp.Value)
	}

	// Read after close should return EOF
	_, err = stream.Recv(ctx)
	assert.Equal(s.T(), io.EOF, err, "Recv after CloseSend should return EOF")
	_, err = stream.Recv(ctx)
	assert.Equal(s.T(), io.EOF, err, "Recv after CloseSend should return EOF")

	for expected := range replies {
		err := stream.Send(ctx, replies[expected])
		assert.NoError(s.T(), err)
	}

	time.Sleep(100 * time.Millisecond) // Give some time for the server to process the close
}

func (s *BidiStreamTestSuite) TestStreamServerError() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	expected := errors.New("server error")

	// Setup expected call
	s.serverMock.On("Stream", mock.Anything, mock.Anything).Return(expected)

	stream, err := s.client.Stream(ctx)
	assert.NoError(s.T(), err)
	assert.NotNil(s.T(), stream)

	reply, err := stream.Recv(ctx)
	assert.Error(s.T(), err)
	assert.Nil(s.T(), reply)
	assert.Equal(s.T(), expected, err)
}

func TestBidiStreamOverTcp(t *testing.T) {
	// Create two tcp transports
	clientConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := tpc.NewTransport(transport.WithAddress("localhost:8080"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewBidiStreamTestSuite(clientConn, serverConn))
}

func TestBidiStreamOverUnix(t *testing.T) {
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

	suite.Run(t, NewBidiStreamTestSuite(clientConn, serverConn))
}

func TestBidiStreamOverMqtt(t *testing.T) {
	// Create two mqtt transports
	clientConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, clientConn)
	defer clientConn.Close()

	serverConn, err := mqtt.NewTransport(transport.WithAddress("localhost:1883"))
	assert.NoError(t, err)
	assert.NotNil(t, serverConn)
	defer serverConn.Close()

	suite.Run(t, NewBidiStreamTestSuite(clientConn, serverConn))
}
