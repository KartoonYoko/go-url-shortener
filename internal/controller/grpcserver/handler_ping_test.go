package grpcserver

import (
	"context"
	"fmt"
	"testing"

	"github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/mocks"
	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func Test_grpcController_Ping(t *testing.T) {
	ctx := context.Background()

	// устанавливаем соединение с сервером
	conn, err := grpc.NewClient(bootstrapAddressgRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	c := pb.NewPingServiceClient(conn)
	type test struct {
		name            string
		prepare         func(mock *mocks.MockUseCasePinger)
		statusErrorCode codes.Code
	}
	tests := []test{
		{
			name: "Success",
			prepare: func(m *mocks.MockUseCasePinger) {
				m.EXPECT().Ping(gomock.Any()).Return(nil)
			},
		},
		{
			name: "Error",
			prepare: func(m *mocks.MockUseCasePinger) {
				m.EXPECT().Ping(gomock.Any()).Return(fmt.Errorf("some unexpected error"))
			},
			statusErrorCode: codes.Internal,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			m := mocks.NewMockUseCasePinger(ctrl)

			if tt.prepare != nil {
				tt.prepare(m)
			}

			controller.ucPing = m

			request := new(pb.PingRequest)
			_, err := c.Ping(ctx, request)

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
