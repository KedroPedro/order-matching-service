package usecases

import (
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/engine"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/interfaces"
)

type AddOrderUsecase struct {
	engine     *engine.Engine
	marketRepo interfaces.MarketStateRepository
}

func NewAddOrderUsecase(
	engine *engine.Engine,
	marketRepo interfaces.MarketStateRepository,
) *AddOrderUsecase {
	return &AddOrderUsecase{
		engine:     engine,
		marketRepo: marketRepo,
	}
}

func (this *AddOrderUsecase) Execute(order *entity.Order) {
	this.engine.AddOrder(order)
}
