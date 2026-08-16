package usecases

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/engine"

type StartDayUsecase struct {
	engine *engine.Engine
}

func NewStartDayUsecase(engine *engine.Engine) *StartDayUsecase {
	return &StartDayUsecase{
		engine: engine,
	}
}

func (this *StartDayUsecase) Execute() {
	this.engine.Start()
}
