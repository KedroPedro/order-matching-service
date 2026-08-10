package usecases

import (
	"context"
	"errors"

	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
)

type CancelOrderUsecase struct {
	marketRepo interfaces.MarketStateRepository
	engine     interfaces.Engine
}

func NewCancelOrderUsecase(
	engine interfaces.Engine,
	marketRepo interfaces.MarketStateRepository,
) *CancelOrderUsecase {
	return &CancelOrderUsecase{
		engine:     engine,
		marketRepo: marketRepo,
	}
}

func (this *CancelOrderUsecase) Execute(ctx context.Context, orderId, ownerId string) error {
	order, err := this.marketRepo.GetById(ctx, orderId)
	if err != nil {
		return err
	}

	if ownerId != order.OwnerId {
		return errs.NewAppError("cancel order error", errors.New("incorrect owner id"))
	}

	this.engine.Cancel(orderId)

	return nil
}
