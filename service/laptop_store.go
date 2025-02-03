package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/copier"
	"github.com/pokala15/pcbook/pb"
)

var ErrAlreadyExists = errors.New("record already exists")

type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	FindById(id string) (*pb.Laptop, error)
	Search(filter *pb.Filter, ctx context.Context, found func(laptop *pb.Laptop) error) error
}

type InMemoryLaptopStore struct {
	mutex sync.RWMutex
	data  map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other, err := createDeepCopy(laptop)
	if err != nil {
		return err
	}

	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) FindById(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if val, ok := store.data[id]; !ok {
		return nil, fmt.Errorf("id not found")
	} else {
		other, err := createDeepCopy(val)
		if err != nil {
			return nil, err
		}
		return other, nil
	}
}

func createDeepCopy(val *pb.Laptop) (*pb.Laptop, error) {
	other := &pb.Laptop{}
	if err := copier.Copy(other, val); err != nil {
		return nil, fmt.Errorf("can't copy the value: %s", err)
	}
	return other, nil
}

func (store *InMemoryLaptopStore) Search(filter *pb.Filter,
	ctx context.Context,
	found func(laptop *pb.Laptop) error,
) error {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, v := range store.data {
		// time.Sleep(time.Second)
		if ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled {
			log.Println("context is cancelled")
			return fmt.Errorf("context is cancelled")
		}
		if isQualifiedLaptop(v, filter) {
			other, err := createDeepCopy(v)
			if err != nil {
				return err
			}
			found(other)
		}
	}
	return nil
}

func isQualifiedLaptop(laptop *pb.Laptop, filter *pb.Filter) bool {
	if laptop.PriceUsd > filter.MaxPriceUsd {
		return false
	}
	if laptop.Cpu.NumberCores < filter.MinCpuCores {
		return false
	}
	if laptop.Cpu.MinGhz < filter.MinCpuGhz {
		return false
	}
	if toBit(laptop.Ram) < toBit(filter.MinRam) {
		return false
	}
	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.Value
	switch memory.Unit {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 8 * 10^3 == 2^3 * 2^10
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}
