package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/pokala15/pcbook/pb"
	"github.com/pokala15/pcbook/service"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("server started on port: %v", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	laptopServer := service.NewLaptopServer(laptopStore, imageStore)
	grpcServer := grpc.NewServer()

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	address := fmt.Sprintf("0.0.0.0:%v", *port)
	listener, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatalf("can't start the server on port: %v", *port)
	}
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("can't start the server on port: %v", *port)
	}
}
