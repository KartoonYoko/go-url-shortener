package grpcserver

import (
	"context"

	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *grpcController) Ping(ctx context.Context, r *pb.PingRequest) (*pb.PingResponse, error) {
	if err := c.ucPing.Ping(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return new(pb.PingResponse), nil
}
