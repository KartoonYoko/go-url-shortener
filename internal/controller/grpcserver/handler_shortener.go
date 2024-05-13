package grpcserver

import (
	"context"

	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	modelShortener "github.com/KartoonYoko/go-url-shortener/internal/model/shortener"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *grpcController) SetURL(ctx context.Context, r *pb.SetURLRequest) (*pb.SetURLResponse, error) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	shortURL, err := c.uc.SaveURL(ctx, r.Url, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	response := new(pb.SetURLResponse)
	response.ShortUrl = shortURL

	return response, nil
}

func (c *grpcController) GetURL(ctx context.Context, r *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	shortURL, err := c.uc.GetURLByID(ctx, r.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	response := new(pb.GetURLResponse)
	response.Url = shortURL

	return response, nil
}

func (c *grpcController) SetURLsBatch(ctx context.Context, r *pb.SetURLsBatchRequest) (*pb.SetURLsBatchResponse, error) {
	userID, err := c.getUserIDFromContext(ctx)
	if err != nil {
		logger.Log.Error("can not get user ID: ", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	request := make([]modelShortener.CreateShortenURLBatchItemRequest, 0, len(r.Items))
	for _, item := range r.Items {
		request = append(request, modelShortener.CreateShortenURLBatchItemRequest{
			OriginalURL:   item.OriginalUrl,
			CorrelationID: item.CorrelationId,
		})
	}
	response, err := c.uc.SaveURLsBatch(ctx, request, userID)
	if err != nil {
		logger.Log.Error("can not save URLs: ", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	pbResponse := new(pb.SetURLsBatchResponse)
	for _, item := range response {
		newItem := &pb.SetURLsBatchResponse_SetURLsBatchResponseItem{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		}

		pbResponse.Items = append(pbResponse.Items, newItem)
	}

	return pbResponse, nil
}
