package grpcserver

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/KartoonYoko/go-url-shortener/config"
	"github.com/google/uuid"
)

var (
	controller           *grpcController
	bootstrapAddressgRPC string
)

// auther - это реализация useCaseAuther, которая позволяет вызывать GetNewUserID сколь угодно раз
type auther struct{}

func (a *auther) GetNewUserID(ctx context.Context) (string, error) {
	id := uuid.New()
	return id.String(), nil
}

func TestMain(m *testing.M) {
	bootstrapAddressgRPC = ":8080"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		fmt.Println("start server")
		controller = createTestMock()
		err := controller.Serve(ctx)
		if err != nil {
			log.Fatalf("server start error: %v", err)
		}

		fmt.Println("stop server")
	}()

	// ждём пока запуститься сервер
	time.Sleep(2 * time.Second)
	m.Run()
}

// createTestMock собирает контроллер
func createTestMock() *grpcController {
	conf := &config.Config{
		BootstrapAddressgRPC: bootstrapAddressgRPC,
		BaseURLAddress:       "http://localhost:8080",
	}
	auther := new(auther)
	c := NewGRPCController(conf, nil, nil, auther, nil)

	return c
}
