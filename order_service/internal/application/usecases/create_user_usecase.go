package usecases

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
)

type CreateUserUsecase struct {
	userRepo interfaces.UserRepository
}

func NewCreateUserUsecase(userRepo interfaces.UserRepository) *CreateUserUsecase {
	return &CreateUserUsecase{
		userRepo: userRepo,
	}
}

func (this *CreateUserUsecase) Execute(ctx context.Context, user *entity.User) error {
	if err := this.userRepo.CreateUser(ctx, user); err != nil {
		return errs.NewAppError("create user error", err)
	}

	return nil
}
