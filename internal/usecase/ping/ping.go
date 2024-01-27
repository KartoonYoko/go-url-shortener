package ping

import "context"

type PingRepo interface {
	Ping(ctx context.Context) error
}

type pingUsecase struct {
	repository PingRepo
}

func NewPingUseCase(r PingRepo) *pingUsecase {
	return &pingUsecase{
		repository: r,
	}
}

func (uc *pingUsecase) Ping(ctx context.Context) error {
	return uc.repository.Ping(ctx)
}
