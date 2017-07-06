package main

import (
	"io"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/ericsage/cxmate/proto"
)

// TestNewServiceConn checks that a connection to a real grpc server can be established and torn down correctly.
func TestNewServiceConn(t *testing.T) {
	time.Sleep(time.Millisecond)
	conn, err := NewServiceConn("0.0.0.0:8080")
	if err != nil {
		t.Fatal(err)
	}
	err = conn.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// TestSendElements checks that sent elements to a mock stream are received.
func TestSendElements(t *testing.T) {
	eles := make(chan *proto.NetworkElement, 10)
	closed := make(chan error)
	count := 0
	mock := StreamMock{eles, closed, &count}
	ss := &ServiceStream{mock}
	c := make(chan *Message)
	ss.OpenSend(c)
	m := &proto.NetworkElement{}
	SendMessage(m, c)
	SendMessage(m, c)
	SendMessage(m, c)
	close(c)
	<-mock.closed
	if len(mock.eles) != 3 {
		t.Error("expected three elements on channel, found", len(mock.eles))
	}
}

// TestReceiveElements checks that elements are receivable from a mock stream.
func TestReceiveElements(t *testing.T) {
	count := 10
	eles := make(chan *proto.NetworkElement, 10)
	mock := StreamMock{eles, nil, &count}
	ss := &ServiceStream{mock}
	c := make(chan *Message)
	ss.OpenReceive(c)
	check := 10
	for _ = range c {
		check--
	}
	if check != 0 {
		t.Fatal("expected a count of 0, found", check)
	}
}

// StreamMock is used for mocking a grc stream
type StreamMock struct {
	eles   chan *proto.NetworkElement
	closed chan error
	count  *int
}

func (s StreamMock) Send(e *proto.NetworkElement) error {
	s.eles <- e
	return nil
}

func (s StreamMock) Recv() (*proto.NetworkElement, error) {
	if *s.count != 0 {
		*s.count--
		return &proto.NetworkElement{}, nil
	}
	return nil, io.EOF
}

func (s StreamMock) CloseSend() error {
	close(s.closed)
	return nil
}

// cxMateServiceServer mock is used to run a test grpc server instance
type cxMateServiceServerMock struct{}

func (s *cxMateServiceServerMock) StreamNetworks(stream proto.CxMateService_StreamNetworksServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = stream.Send(in)
		if err != nil {
			return err
		}
	}
	return nil
}

func runMockServer(t *testing.T, address string) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterCxMateServiceServer(grpcServer, &cxMateServiceServerMock{})
	grpcServer.Serve(lis)
}
