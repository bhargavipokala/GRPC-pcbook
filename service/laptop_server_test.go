package service

import (
	"context"
	"testing"

	"github.com/pokala15/pcbook/pb"
	"github.com/pokala15/pcbook/sample"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopNoId := sample.NewLaptop()
	laptopNoId.Id = ""

	laptopInvalidId := sample.NewLaptop()
	laptopInvalidId.Id = "abc"

	duplicateStore := NewInMemoryLaptopStore()
	laptop := sample.NewLaptop()
	err := duplicateStore.Save(laptop)
	require.Nil(t, err)
	duplicateLaptop := sample.NewLaptop()
	duplicateLaptop.Id = laptop.Id

	testCases := []struct {
		name   string
		laptop *pb.Laptop
		store  LaptopStore
		code   codes.Code
	}{
		{
			name:   "success_with_id",
			laptop: sample.NewLaptop(),
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "success_with_id",
			laptop: laptopNoId,
			store:  NewInMemoryLaptopStore(),
			code:   codes.OK,
		},
		{
			name:   "failure_with_invalid_id",
			laptop: laptopInvalidId,
			store:  NewInMemoryLaptopStore(),
			code:   codes.InvalidArgument,
		},
		{
			name:   "failure_with_dup_id",
			laptop: duplicateLaptop,
			store:  duplicateStore,
			code:   codes.AlreadyExists,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			response, err := NewLaptopServer(tc.store, nil).CreateLaptop(
				context.Background(),
				&pb.CreateLaptopRequest{
					Laptop: tc.laptop,
				},
			)
			if tc.code == codes.OK {
				require.NoError(t, err)
				require.NotNil(t, response)
				require.NotEmpty(t, response.Id)
				require.Equal(t, tc.laptop.Id, response.Id)
			} else {
				require.Error(t, err)
				require.Nil(t, response)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, st.Code(), tc.code)
			}
		})
	}
}
