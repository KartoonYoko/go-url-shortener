/*
Package ping это usecase для проверки работы выбранного хранилища
*/
package ping

import "context"

// PingRepo Интерфейс хранилища
type PingRepo interface {
	Ping(ctx context.Context) error
}

type pingUsecase struct {
	repository PingRepo
}

// NewPingUseCase конструктор pingUsecase
func NewPingUseCase(r PingRepo) *pingUsecase {
	return &pingUsecase{
		repository: r,
	}
}

// Ping реализует Pinger
func (uc *pingUsecase) Ping(ctx context.Context) error {
	return uc.repository.Ping(ctx)
}
