package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/ericsage/mate/proto"
	"google.golang.org/grpc"
)

const address = "0.0.0.0:8080"

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Could not establish connection to cx server.")
	}
	client := proto.NewCxMateServiceClient(conn)
	stream, err := client.StreamNetworks(context.Background())
	if err != nil {
		log.Fatal("Could not call StreamNetworks function with client connection to server.")
	}

	ele := &proto.NetworkElement{
		NetworkId: 0,
		Element: &proto.NetworkElement_Parameter{
			Parameter: &proto.Parameter{
				Name:  "test_name",
				Value: "test_value",
			},
		},
	}
	log.Println("Sending parameter element...")
	err = stream.Send(ele)
	log.Println("Parameter element sent... checking for errors.")
	if err != nil {
		log.Printf("Send error: %#v", err)
	}
	log.Println("Closing the send channel...")
	err = stream.CloseSend()
	if err != nil {
		log.Printf("CloseSend error: %#v", err)
	}
	log.Println("Listening for incomming messages.")
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Recv error: %#v", err)
			break
		}
		fmt.Printf("Received element: \n %#v", in)
	}
}
