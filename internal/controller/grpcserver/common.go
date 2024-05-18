package grpcserver

import (
	"context"
	"fmt"

	"github.com/KartoonYoko/go-url-shortener/internal/logger"
)

func (c *grpcController) getUserIDFromContext(ctx context.Context) (string, error) {
	ctxUserID := ctx.Value(keyUserID)
	userID, ok := ctxUserID.(string)
	if !ok {
		msg := "can not get user ID from context"
		logger.Log.Debug(msg)
		return "", fmt.Errorf(msg)
	}

	return userID, nil
}
