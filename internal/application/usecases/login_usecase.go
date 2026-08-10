package usecases

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/rs/zerolog/log"
)

type LoginUsecase struct {
	userRepo    interfaces.UserRepository
	sessionRepo interfaces.SessionRepository
}

func NewLoginUsecase(userRepo interfaces.UserRepository, sessionRepo interfaces.SessionRepository) *LoginUsecase {
	return &LoginUsecase{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (this *LoginUsecase) Execute(ctx context.Context, login string) (*entity.User, error) {
	user, err := this.sessionRepo.GetSession(ctx, login)
	if err == nil {
		return user, nil
	}

	user, err = this.userRepo.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, errs.NewAppError("login user error", err)
	}

	go func() {
		if err := this.sessionRepo.AddSession(ctx, user); err != nil {
			log.Err(errs.NewAppError("save session error", err)).Send()
		}
	}()

	return user, nil
}
