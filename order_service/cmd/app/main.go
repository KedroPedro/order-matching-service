package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/controller"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/engine"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/mongodb"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/redis"
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
	sessionRepo := redisClient.NewSessionRepository()

	appWg := sync.WaitGroup{}

	engineClient := engine.NewClient()

	eng, err := engineClient.NewEngine()
	if err != nil {
		panic(err)
	}

	router := controller.NewRouter(
		appCtx,
		marketRepo,
		eng,
		orderRepo,
		userRepo,
		sessionRepo,
	)

	if err := SetupServer(router); err != nil {
		panic(err)
	}

	appWg.Wait()

}
