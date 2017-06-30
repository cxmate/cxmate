package main

import (
	"io"
	"testing"

	"github.com/ericsage/cxmate/proto"
)

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
