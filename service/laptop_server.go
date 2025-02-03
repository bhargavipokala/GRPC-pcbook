package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/pokala15/pcbook/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxImageSize = 1 << 20

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	laptopStore LaptopStore
	imageStore  ImageStore
}

func NewLaptopServer(store LaptopStore, imageStore ImageStore) *LaptopServer {
	return &LaptopServer{
		laptopStore: store,
		imageStore:  imageStore,
	}
}

func (service *LaptopServer) CreateLaptop(
	ctx context.Context,
	request *pb.CreateLaptopRequest,
) (response *pb.CreateLaptopResponse, err error) {
	laptop := request.GetLaptop()
	log.Printf("receive create laptop request with id: %s", laptop.Id)

	if len(laptop.Id) > 0 {
		if err := uuid.Validate(laptop.Id); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptopId is not a valid uuid: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to created id for laptop: %v", err)
		}
		laptop.Id = id.String()
	}

	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	// store the laptop object in storage
	if err := service.laptopStore.Save(laptop); err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "failed to save laptop: %v", err)
	} else {
		log.Printf("laptop is successfully saved with id: %v", laptop.Id)
	}
	return &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}, nil
}

func validateContext(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		const message = "request is cancelled"
		log.Println(message)
		return status.Error(codes.Canceled, message)
	case context.DeadlineExceeded:
		const message = "deadline is exceeded"
		log.Println(message)
		return status.Error(codes.DeadlineExceeded, message)
	default:
		return nil
	}
}

func (service *LaptopServer) SearchLaptop(request *pb.SearchLaptopRequest,
	stream grpc.ServerStreamingServer[pb.SearchLaptopResponse],
) error {
	filter := request.GetFilter()
	log.Printf("receive search laptop with filter: %v", filter)

	err := service.laptopStore.Search(filter, stream.Context(),
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}
			err := stream.Send(res)

			if err != nil {
				return err
			}

			log.Printf("sent laptop to client with id: %v", res.Laptop.Id)
			return nil
		})
	if err != nil {
		return status.Errorf(codes.Internal, "error while searching for laptop: %v", err)
	}
	return nil
}

func (service *LaptopServer) UploadImage(stream grpc.ClientStreamingServer[pb.UploadImageRequest,
	pb.UploadImageResponse]) error {
	imageData := bytes.Buffer{}
	imageSize := 0

	request, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.Unknown, "failed to read streaming data: %v", err)
	}
	laptopId := request.GetInfo().GetLaptopId()
	imageType := request.GetInfo().GetImageType()

	laptop, err := service.laptopStore.FindById(laptopId)
	if err != nil {
		return status.Errorf(codes.Internal, "error while fetching laptop: %v", err)
	} else if laptop == nil {
		return status.Errorf(codes.InvalidArgument, "laptop doesn't exist with id: %v", laptopId)
	}

	for {
		request, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return status.Errorf(codes.Unknown, "failed to read streaming data: %v", err)
		}
		if err := validateContext(stream.Context()); err != nil {
			return err
		}
		chunk := request.GetChunkData()
		chunkSize, err := imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Unknown, "unable to write the image to file: %v", err)
		}
		imageSize += chunkSize
		if imageSize > maxImageSize {
			return status.Errorf(codes.Unknown, "image is too big : %v > %v", imageSize, maxImageSize)
		}
	}

	imageId, err := service.imageStore.Save(laptopId, imageType, imageData)
	if err != nil {
		return status.Errorf(codes.Internal, "Unable save the image: %v", err)
	}

	return stream.SendAndClose(&pb.UploadImageResponse{
		ImageId: imageId,
		Size:    uint32(imageSize),
	})
}
