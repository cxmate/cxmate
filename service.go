package main

import (
	"context"
	"errors"
	"io"

	"github.com/ericsage/cxmate/proto"
	"google.golang.org/grpc"
)

// ServiceConn holds a connect to the service that should persistsfor the lifetime of cxMate
type ServiceConn struct {
	conn *grpc.ClientConn
}

// NewServiceConn creates a persistent connection to the service at an address
func NewServiceConn(address string) (*ServiceConn, error) {
	logDebugln("Dialing service...")
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, errors.New("could not establish connection")
	}
	logDebugln("Connection to service established")
	return &ServiceConn{conn: conn}, nil
}

// Close tears down the ServiceConn connection, including all active streams
func (sc *ServiceConn) Close() error {
	logDebugln("Closing connection to service")
	return sc.conn.Close()
}

//Stream allows mocking of a concrete grpc stream client
type Stream interface {
	Send(e *proto.NetworkElement) error
	Recv() (*proto.NetworkElement, error)
	CloseSend() error
}

// ServiceStream is a bidirectional stream between cxmate and the service that services a single client request. Use OpenSend and
// OpenRecieve to create a channel per streaming direction
type ServiceStream struct {
	stream Stream
}

// NewServiceStream creates a client for communicating with a service
func (sc *ServiceConn) NewServiceStream() (*ServiceStream, error) {
	logDebugln("Creating new service stream to process request")
	c := proto.NewCxMateServiceClient(sc.conn)
	stream, err := c.StreamNetworks(context.Background())
	if err != nil {
		return nil, errors.New("could not initiate StreamNetworks call to server")
	}
	logDebugln("Service stream initialized")
	return &ServiceStream{stream}, nil
}

// Message contains an element to be sent to the service and a channel for detecting errors during the sending
type Message struct {
	ele     *proto.NetworkElement
	errChan chan error
	err     error
}

// OpenSend spawns a go routine that sends wrapped elements placed on a channel to the service, and may place an error per element into an error channel
// sent in the wrapper, and will always close the error channel after the element is sent. Closes the rpc stream to the service
// and returns when the channel is clsoed
func (ss *ServiceStream) OpenSend(s <-chan *Message) {
	go func() {
		logDebugln("Initiating read from the send message channel")
		for sm := range s {
			err := ss.stream.Send(sm.ele)
			if err != nil {
				sm.errChan <- err
			}
			close(sm.errChan)
		}
		logDebugln("Closing the send message channel")
		ss.stream.CloseSend()
	}()
}

// SendMessage sends an element to a service and blocks until an error value (may be nil) is received
func SendMessage(ele *proto.NetworkElement, s chan<- *Message) error {
	errChan := make(chan error)
	s <- &Message{
		ele:     ele,
		errChan: errChan,
	}
	err := <-errChan
	return err
}

// OpenReceive spawns a go routine that listens for elements sent by the service, and puts element, error messages into a channel. Closes the channel and returns
// when the service closes the stream (or disconnects erroneously)
func (ss *ServiceStream) OpenReceive(r chan<- *Message) {
	go func() {
		logDebugln("Initiating send to the receive message channel")
		for {
			ele, err := ss.stream.Recv()
			if err == io.EOF {
				break
			}
			r <- &Message{
				ele: ele,
				err: err,
			}
		}
		logDebugln("Closing the receive message channel")
		close(r)
	}()
}

// ReceiveMessage blocks until it recieves an element from the service, returns the element or an error. error is set to io.EOF if the
// channel is closed (indicating the service is finished sending (or disconnected erroneously))
func ReceiveMessage(r <-chan *Message) (*proto.NetworkElement, error) {
	m, ok := <-r
	if !ok {
		return nil, io.EOF
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.ele, nil
}
