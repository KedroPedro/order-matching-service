package usecases

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/engine"

type EndDayUsecase struct {
	engine *engine.Engine
}

func NewEndDayUsecase(engine *engine.Engine) *EndDayUsecase {
	return &EndDayUsecase{
		engine: engine,
	}
}

func (this *EndDayUsecase) Execute() {

	this.engine.Stop()
	this.engine.EndDay()
}
