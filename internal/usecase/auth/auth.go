/*
Package auth это usecase для работы с аутентификацией пользователя
*/
package auth

import "context"

// AuthRepo интерфейс для хранилища
type AuthRepo interface {
	GetNewUserID(ctx context.Context) (string, error)
}

type authUseCase struct {
	repository AuthRepo
}

// NewAuthUseCase конструктор authUseCase
func NewAuthUseCase(r AuthRepo) *authUseCase {
	return &authUseCase{
		repository: r,
	}
}

// GetNewUserID вернёт уникальный ID пользователя
func (uc *authUseCase) GetNewUserID(ctx context.Context) (string, error) {
	return uc.repository.GetNewUserID(ctx)
}
