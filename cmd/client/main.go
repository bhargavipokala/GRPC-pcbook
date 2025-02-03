package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/pokala15/pcbook/pb"
	"github.com/pokala15/pcbook/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	address := flag.String("address", "", "conn address")
	flag.Parse()

	conn, err := grpc.NewClient(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to create connection to address: %s", *address)
	}

	client := pb.NewLaptopServiceClient(conn)

	for i := 0; i <= 10; i++ {
		createLaptop(client, sample.NewLaptop())
	}
	searchLaptop(client)
	uploadImage(client)
}

func uploadImage(client pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(client, laptop)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.UploadImage(ctx)
	if err != nil {
		log.Fatalf("unable to upload the image: %v", err)
	}

	err = stream.Send(&pb.UploadImageRequest{
		Info: &pb.ImageInfo{
			LaptopId:  laptop.Id,
			ImageType: pb.ImageType_JPG,
		},
	})
	if err != nil {
		log.Fatalf("unable to send the info: %v", err)
	}

	file, err := os.Open("tmp/laptop.jpg")
	if err != nil {
		log.Fatalf("error while opening the file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("unable to read the file: %v", err)
		}
		err = stream.Send(&pb.UploadImageRequest{
			ChunkData: buffer[:n],
		})
		if err != nil {
			log.Fatalf("unable to send the info: %v", err)
		}
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error while receiving the response: %v", err)
	}
	fmt.Printf("uploaded image with id %v of size %v", response.GetImageId(), response.GetSize())
}

func searchLaptop(client pb.LaptopServiceClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := &pb.Filter{
		MaxPriceUsd: 3000,
		MinCpuCores: 3,
		MinCpuGhz:   1,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	stream, err := client.SearchLaptop(ctx, &pb.SearchLaptopRequest{
		Filter: filter,
	})

	if err != nil {
		log.Fatalf("couldn't search laptop: %v", err)
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			return
		} else if err != nil {
			log.Fatalf("couldn't recieve response: %v", err)
		}

		laptop := response.GetLaptop()
		log.Printf("recieved laptop with id: %v\n", laptop.Id)
	}
}

func createLaptop(client pb.LaptopServiceClient, laptop *pb.Laptop) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := client.CreateLaptop(ctx, &pb.CreateLaptopRequest{
		Laptop: laptop,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Printf("Laptop is created with id: %v", response.Id)
		} else {
			log.Fatalf("couldn't create laptop: %s", err)
		}
	} else {
		log.Printf("Laptop is created with id: %v", response.Id)
	}
}
