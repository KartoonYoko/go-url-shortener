package grpcserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/KartoonYoko/go-url-shortener/config"
	pb "github.com/KartoonYoko/go-url-shortener/internal/controller/grpcserver/proto"
	"github.com/KartoonYoko/go-url-shortener/internal/logger"
	"google.golang.org/grpc"
)

type grpcController struct {
	uc      UseCaseShortener
	ucPing  UseCasePinger
	ucAuth  useCaseAuther
	ucStats UseCaseStats

	pb.PingServiceServer
	pb.StatsServiceServer
	pb.ShortenerServiceServer

	conf *config.Config
}

func NewGRPCController(
	conf *config.Config,
	uc UseCaseShortener,
	ucPing UseCasePinger,
	ucAuth useCaseAuther,
	ucStats UseCaseStats) *grpcController {
	c := new(grpcController)
	c.conf = conf
	c.uc = uc
	c.ucAuth = ucAuth
	c.ucPing = ucPing
	c.ucStats = ucStats

	return c
}

func (c *grpcController) Serve(ctx context.Context) error {
	_, cancel := context.WithCancel(ctx)
	defer cancel()

	listen, err := net.Listen("tcp", c.conf.BootstrapAddressgRPC)
	if err != nil {
		return fmt.Errorf("failed to start grpc server: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		c.interceptorRequestTime,
		c.interceptorAuth,
	))
	pb.RegisterPingServiceServer(grpcServer, c)
	pb.RegisterStatsServiceServer(grpcServer, c)
	pb.RegisterShortenerServiceServer(grpcServer, c)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		s := <-sigCh
		logger.Log.Info(fmt.Sprintf("got signal %v, attempting graceful shutdown", s))
		cancel()

		grpcServer.GracefulStop()
		wg.Done()
	}()

	logger.Log.Info(fmt.Sprintf("grpc serve on %s", c.conf.BootstrapAddressgRPC))
	if err := grpcServer.Serve(listen); err != nil {
		return fmt.Errorf("serve error grpc server: %w", err)
	}
	wg.Wait()
	logger.Log.Info("grpc server stopped gracefully")

	return nil
}
