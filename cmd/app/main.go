package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	eventhandler "github.com/KedroPedro/order-matching-engine/internal/application/event_handler"
	"github.com/KedroPedro/order-matching-engine/internal/controller"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/starter"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/mongodb"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/redis"
	"github.com/rs/zerolog/log"
)

func main() {
	appCtx, cancel := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt)
	defer cancel()

	mongoClient, err := mongodb.NewMongoClient()
	if err != nil {
		panic(err)
	}

	redisClient, err := redis.NewClient()
	if err != nil {
		panic(err)
	}

	marketRepo := redisClient.NewMarketRepository()
	if err := marketRepo.LoadScripts(appCtx); err != nil {
		panic(err)
	}

	orderRepo := mongoClient.NewOrderRepository(appCtx)
	userRepo := mongoClient.NewUserRepository()

	appWg := sync.WaitGroup{}

	engineClient, eventCh := starter.New(appCtx)

	appWg.Go(func() {
		engineClient.Start(appCtx)
		log.Debug().Msg("engine stopped")
	})

	eventHandler := eventhandler.NewEventHandler(
		eventCh,
		marketRepo,
		orderRepo,
	)

	appWg.Go(func() {
		eventHandler.Start(appCtx)
		log.Debug().Msg("event handler stopped")
	})

	router := controller.NewRouter(
		appCtx,
		marketRepo,
		engineClient,
		orderRepo,
		userRepo,
	)

	if err := SetupServer(router); err != nil {
		panic(err)
	}

	appWg.Wait()

}
