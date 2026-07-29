package controller

import (
	"context"
	"net/http"

	"github.com/KedroPedro/order-matching-engine/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/internal/controller/handlers"
	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
)

type Router struct {
	marketHandler *handlers.MarketStateHandler
	ordersHandler *handlers.OrdersHandler
	usersHandler  *handlers.UsersHandler
	mux           *http.ServeMux
}

func NewRouter(
	ctx context.Context,
	marketRepo interfaces.MarketStateRepository,
	engine interfaces.Engine,
	orderRepo interfaces.OrderRepository,
	userRepo interfaces.UserRepository,
) *Router {
	getStateUsecase := usecases.NewGetStateUsecase(marketRepo)
	addOrderUsecase := usecases.NewAddOrderUsecase(orderRepo, userRepo, marketRepo, engine)
	cancelOrderUsecase := usecases.NewCancelOrderUsecase(engine, marketRepo)

	mux := http.ServeMux{}

	r := &Router{
		marketHandler: handlers.NewMarketStateHandler(ctx, getStateUsecase, &mux),
		ordersHandler: handlers.NewOrdersHandler(addOrderUsecase, cancelOrderUsecase, &mux),
		usersHandler:  handlers.NewUsersHandler(&mux, usecases.NewCreateUserUsecase(userRepo), usecases.NewLoginUsecase(userRepo)),
		mux:           &mux,
	}

	return r
}

func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.mux.ServeHTTP(w, r)
}
