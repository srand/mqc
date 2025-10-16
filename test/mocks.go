package test

import (
	"github.com/srand/mqc"
	"github.com/stretchr/testify/mock"
)

type RpcTestServerMock struct {
	mock.Mock
}

func (s *RpcTestServerMock) Rpc(req *TestRequest) (*TestReply, error) {
	args := s.Called(req)
	reply := args.Get(0)
	if reply == nil {
		return nil, args.Error(1)
	}
	return reply.(*TestReply), args.Error(1)
}

type ServerStreamTestServerMock struct {
	mock.Mock
}

func (s *ServerStreamTestServerMock) Stream(req *TestRequest, stream mqc.ServerStreamServer[TestReply]) error {
	args := s.Called(req, stream)
	return args.Error(0)
}

type ClientStreamTestServerMock struct {
	mock.Mock
}

func (s *ClientStreamTestServerMock) Stream(stream mqc.ClientStreamServer[TestRequest, TestReply]) error {
	args := s.Called(stream)
	return args.Error(0)
}

type BidiStreamTestServerMock struct {
	mock.Mock
}

func (s *BidiStreamTestServerMock) Stream(stream mqc.BidiStreamServer[TestRequest, TestReply]) error {
	args := s.Called(stream)
	return args.Error(0)
}
