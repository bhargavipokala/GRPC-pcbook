package service

import (
	"context"
	"io"
	"net"
	"testing"

	"github.com/pokala15/pcbook/pb"
	"github.com/pokala15/pcbook/sample"
	"github.com/pokala15/pcbook/serializer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := NewInMemoryLaptopStore()
	_, serverAdd := startTestLaptopServer(t, laptopStore, nil)
	laptopClient := newTestLaptopClient(t, serverAdd)

	laptop := sample.NewLaptop()
	expectedId := laptop.Id

	response, err := laptopClient.CreateLaptop(context.Background(), &pb.CreateLaptopRequest{
		Laptop: laptop,
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, expectedId, response.Id)

	// check for the object saved
	savedLaptop, err := laptopStore.FindById(response.Id)
	require.NoError(t, err)
	requireSameLaptop(t, laptop, savedLaptop)
}

func TestClientSearchLaptop(t *testing.T) {
	store := NewInMemoryLaptopStore()
	expectedIds := make(map[string]bool)

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinRam:      &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 8},
	}

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()
		switch i {
		case 0:
			laptop.PriceUsd = 2200
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2
		case 3:
			laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 4}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 16}
			expectedIds[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.5
			laptop.Ram = &pb.Memory{Unit: pb.Memory_GIGABYTE, Value: 16}
			expectedIds[laptop.Id] = true
		}
		err := store.Save(laptop)
		require.NoError(t, err)
	}
	_, serverAdd := startTestLaptopServer(t, store, nil)
	laptopClient := newTestLaptopClient(t, serverAdd)

	stream, err := laptopClient.SearchLaptop(context.Background(), &pb.SearchLaptopRequest{
		Filter: filter,
	})
	require.NoError(t, err)
	found := 0
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		require.Contains(t, expectedIds, response.GetLaptop().GetId())
		found += 1
	}
	require.Equal(t, len(expectedIds), found)
}

func startTestLaptopServer(t *testing.T, store LaptopStore, imageStore ImageStore) (*LaptopServer, string) {
	laptopServer := NewLaptopServer(store, imageStore)

	grpcServer := grpc.NewServer()
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	listener, err := net.Listen("tcp", ":0")
	require.Nil(t, err)

	go grpcServer.Serve(listener)

	return laptopServer, listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAdd string) pb.LaptopServiceClient {
	conn, err := grpc.NewClient(serverAdd, grpc.WithInsecure())
	require.Nil(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func requireSameLaptop(t *testing.T, laptop1 *pb.Laptop, laptop2 *pb.Laptop) {
	laptop1Json := serializer.ProtobufToJSON(laptop1)
	laptop2Json := serializer.ProtobufToJSON(laptop2)
	require.Equal(t, laptop1Json, laptop2Json)
}
