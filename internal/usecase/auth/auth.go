package auth

import "context"

type AuthRepo interface {
	GetNewUserID(ctx context.Context) (string, error)
}

type authUseCase struct {
	repository AuthRepo
}

func NewAuthUseCase(r AuthRepo) *authUseCase {
	return &authUseCase{
		repository: r,
	}
}

func (uc *authUseCase) GetNewUserID(ctx context.Context) (string, error) {
	return uc.repository.GetNewUserID(ctx)
}
