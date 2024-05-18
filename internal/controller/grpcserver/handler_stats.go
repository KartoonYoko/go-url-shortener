package grpcserver

import (
	"context"

	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *grpcController) GetStats(ctx context.Context, r *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := c.ucStats.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	res := new(pb.GetStatsResponse)
	res.Urls = int64(stats.URLs)
	res.Users = int64(stats.Users)

	return res, nil
}
