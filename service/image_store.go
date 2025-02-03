package service

import (
	"bytes"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/pokala15/pcbook/pb"
)

type ImageStore interface {
	Save(laptopId string, imageType pb.ImageType, imageData bytes.Buffer) (string, error)
}

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

type ImageInfo struct {
	LaptopId string
	Type     *pb.ImageType
	Path     string
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

func (imageStore *DiskImageStore) Save(laptopId string,
	imageType pb.ImageType, imageData bytes.Buffer) (string, error) {
	imageId, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("error while creating imageId: %v", err)
	}
	imagePath := fmt.Sprintf("%s/%s.%s", imageStore.imageFolder, laptopId, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("error while creating file: %v", err)
	}
	defer file.Close()
	_, err = imageData.WriteTo(file)
	if err != nil {
		return "", fmt.Errorf("error while writing to file: %v", err)
	}

	imageStore.mutex.Lock()
	defer imageStore.mutex.Unlock()

	imageStore.images[imageId.String()] = &ImageInfo{
		LaptopId: laptopId,
		Type:     &imageType,
		Path:     imagePath,
	}

	return imageId.String(), nil
}
