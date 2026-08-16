package usecases

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
)

const (
	safetyFactor = 1.05
)

type AddOrderUsecase struct {
	orderRepo  interfaces.OrderRepository
	userRepo   interfaces.UserRepository
	marketRepo interfaces.MarketStateRepository
	engine     interfaces.Engine
}

func NewAddOrderUsecase(
	orderRepo interfaces.OrderRepository,
	userRepo interfaces.UserRepository,
	marketRepo interfaces.MarketStateRepository,
	engine interfaces.Engine,
) *AddOrderUsecase {
	return &AddOrderUsecase{
		orderRepo:  orderRepo,
		userRepo:   userRepo,
		marketRepo: marketRepo,
		engine:     engine,
	}
}

func (this *AddOrderUsecase) Execute(ctx context.Context, order *entity.Order) error {
	switch order.Class {
	case entity.Limit:
		order.Reserve = order.Quantity * order.Price

	case entity.Market:
		bestPrice, err := this.marketRepo.GetBestPrice(ctx, order.Type)
		if err != nil {
			return errs.NewAppError("get best price error", err)
		}

		order.Reserve = int64(float64(order.Quantity*bestPrice) * safetyFactor)

	default:
		return errs.NewAppError("add order error", errs.NewTypeError("undefined order class"))
	}

	if err := this.userRepo.ReserveBalance(ctx, order.OwnerId, order.Reserve); err != nil {
		return errs.NewAppError("reserve balance error", err)
	}

	var err error

	defer func() {
		if err != nil {
			this.userRepo.ReleaseBalance(ctx, order.OwnerId, order.Reserve)
		}
	}()

	if err = this.marketRepo.AddOrder(ctx, order); err != nil {
		return errs.NewAppError("add order to market error", err)
	}

	if err = this.orderRepo.AddOrder(ctx, order); err != nil {
		return errs.NewAppError("add order to storage error", err)
	}

	if err = this.engine.AddOrder(ctx, order); err != nil {
		return errs.NewAppError("add order to engine error", err)
	}

	return nil
}
