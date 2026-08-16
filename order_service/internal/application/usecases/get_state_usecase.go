package usecases

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
)

type GetStateUsecase struct {
	marketRepo interfaces.MarketStateRepository
}

func NewGetStateUsecase(marketRepo interfaces.MarketStateRepository) *GetStateUsecase {
	return &GetStateUsecase{
		marketRepo: marketRepo,
	}
}

func (this *GetStateUsecase) Execute(ctx context.Context) (map[string]string, map[string]string, error) {
	asks, bids, err := this.marketRepo.GetState(ctx)
	if err != nil {
		return nil, nil, errs.NewAppError("get market state error", err)
	}

	return asks, bids, nil
}
