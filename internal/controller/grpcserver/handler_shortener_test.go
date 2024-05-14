package grpcserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/mocks"
	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	modelShortener "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Test_grpcController_SetURL(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCaseShortener)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().SaveURL(gomock.Any(), gomock.Any(), gomock.Any()).Return("", fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCaseShortener(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.uc = m

			request := new(pb.SetURLRequest)
			_, err := c.SetURL(ctx, request)

			if tt.statusErrorCode == 0 {
				require.NoError(t, err)
			} else {
				if e, ok := status.FromError(err); ok {
					require.Equal(t, tt.statusErrorCode, e.Code())
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}

func Test_grpcController_GetURL(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCaseShortener)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().GetURLByID(gomock.Any(), gomock.Any()).Return("", nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().GetURLByID(gomock.Any(), gomock.Any()).Return("", fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCaseShortener(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.uc = m

			request := new(pb.GetURLRequest)
			_, err := c.GetURL(ctx, request)

			if tt.statusErrorCode == 0 {
				require.NoError(t, err)
			} else {
				if e, ok := status.FromError(err); ok {
					require.Equal(t, tt.statusErrorCode, e.Code())
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}

func Test_grpcController_SetURLsBatch(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCaseShortener)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCaseShortener) {
				arr := make([]modelShortener.CreateShortenURLBatchItemResponse, 0)
				m.EXPECT().SaveURLsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(arr, nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().SaveURLsBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCaseShortener(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.uc = m

			request := new(pb.SetURLsBatchRequest)
			_, err := c.SetURLsBatch(ctx, request)

			if tt.statusErrorCode == 0 {
				require.NoError(t, err)
			} else {
				if e, ok := status.FromError(err); ok {
					require.Equal(t, tt.statusErrorCode, e.Code())
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}

func Test_grpcController_GetUserURLs(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCaseShortener)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCaseShortener) {
				arr := make([]modelShortener.GetUserURLsItemResponse, 0)
				m.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return(arr, nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().GetUserURLs(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCaseShortener(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.uc = m

			request := new(pb.GetUserURLsRequest)
			_, err := c.GetUserURLs(ctx, request)

			if tt.statusErrorCode == 0 {
				require.NoError(t, err)
			} else {
				if e, ok := status.FromError(err); ok {
					require.Equal(t, tt.statusErrorCode, e.Code())
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}

func Test_grpcController_DeleteUserURLs(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewShortenerServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCaseShortener)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().DeleteURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCaseShortener) {
				m.EXPECT().DeleteURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCaseShortener(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.uc = m

			request := new(pb.DeleteUserURLsRequest)
			_, err := c.DeleteUserURLs(ctx, request)

			if tt.statusErrorCode == 0 {
				require.NoError(t, err)
			} else {
				if e, ok := status.FromError(err); ok {
					require.Equal(t, tt.statusErrorCode, e.Code())
				} else {
					t.Errorf("unexpected error: %v", err)
				}
			}

		})
	}
}
