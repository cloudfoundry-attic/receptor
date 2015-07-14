package main

import (
	"log"

	"github.com/cloudfoundry-incubator/bbs/models/service"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:7777")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := service.NewDesiredLRPServiceClient(conn)

	r, err := c.DesiredLRPs(context.Background(), &service.DesiredRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	println("response", len(r.GetDesiredLRPs()))
}
