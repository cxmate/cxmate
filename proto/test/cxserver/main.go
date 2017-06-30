package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/ericsage/cxmate/proto"
	"google.golang.org/grpc"
)

type cxMateServiceServer struct{}

//StreamElements echos back any elements sent to it by a CyService client
func (s *cxMateServiceServer) StreamNetworks(stream proto.CxMateService_StreamNetworksServer) error {
	t0 := time.Now()
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			log.Println("EOF encountered in err... ending call....")
			break
		}
		if err != nil {
			log.Printf("Recv error: %#v", err)
			return err
		}
		/*
			err = stream.Send(in)
			if err != nil {
				return err
			}
		*/
	}
	t1 := time.Now()
	fmt.Printf("The call took %v to run.\n", t1.Sub(t0))
	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterCxMateServiceServer(grpcServer, &cxMateServiceServer{})
	log.Println("cxecho now listening on 0.0.0.0:8080 for incoming grpc connections.")
	grpcServer.Serve(lis)
}
