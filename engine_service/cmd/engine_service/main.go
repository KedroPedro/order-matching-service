package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	eventhandler "github.com/KedroPedro/order-matching-engine/engine_service/internal/application/event_handler"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/controller/handlers"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/engine"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/infrastructure/mongodb"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/infrastructure/redis"
)

func main() {
	appCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	wg := sync.WaitGroup{}

	redis, err := redis.NewClient()
	if err != nil {
		panic(err)
	}

	marketRepo := redis.NewMarketRepository()
	if err := marketRepo.LoadScripts(appCtx); err != nil {
		panic(err)
	}

	mongo, err := mongodb.NewMongoClient()
	if err != nil {
		panic(err)
	}

	orderRepo := mongo.NewOrderRepository(appCtx)

	eventBatchChan := make(chan *entity.EventBatch, 100000)

	eng := engine.New(eventBatchChan)

	wg.Go(func() {
		eng.Process(appCtx)
	})

	eventHandler := eventhandler.NewEventHandler(eventBatchChan, marketRepo, orderRepo)

	wg.Go(func() {
		eventHandler.Start(appCtx)
	})

	addOrderUsecase := usecases.NewAddOrderUsecase(eng, marketRepo)
	cancelOrderUsecase := usecases.NewCancelOrderUsecase(eng)
	startDayUsecase := usecases.NewStartDayUsecase(eng)
	endDayUsecase := usecases.NewEndDayUsecase(eng)

	comandsHandler := handlers.NewComandsHandler(startDayUsecase, endDayUsecase)

	wg.Go(func() {
		comandsHandler.Handle(appCtx)
	})

	ordersHandler := handlers.NewOrdersHandler(addOrderUsecase, cancelOrderUsecase)

	SetupServer(ordersHandler)

	wg.Wait()
}
