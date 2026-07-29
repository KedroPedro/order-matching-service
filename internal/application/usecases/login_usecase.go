package usecases

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
)

type LoginUsecase struct {
	userRepo interfaces.UserRepository
}

func NewLoginUsecase(userRepo interfaces.UserRepository) *LoginUsecase {
	return &LoginUsecase{
		userRepo: userRepo,
	}
}

func (this *LoginUsecase) Execute(ctx context.Context, login string) (*entity.User, error) {
	user, err := this.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, errs.NewAppError("login user error", err)
	}

	return user, nil
}
