package usecases

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/engine"

type CancelOrderUsecase struct {
	engine *engine.Engine
}

func NewCancelOrderUsecase(engine *engine.Engine) *CancelOrderUsecase {
	return &CancelOrderUsecase{
		engine: engine,
	}
}

func (this CancelOrderUsecase) Execute(id string) {
	this.engine.CancelOrder(id)
}
