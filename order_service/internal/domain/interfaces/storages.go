package interfaces

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
)

type UserRepository interface {
	ReserveBalance(ctx context.Context, userId string, toReserve int64) error
	ReleaseBalance(ctx context.Context, userId string, toReserve int64) error
	CreateUser(ctx context.Context, user *entity.User) error
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
}

type OrderRepository interface {
	AddOrder(ctx context.Context, order *entity.Order) error
}

type MarketStateRepository interface {
	AddOrder(ctx context.Context, order *entity.Order) error
	GetBestPrice(ctx context.Context, orderType entity.OrderType) (int64, error)
	GetState(ctx context.Context) (asks map[string]string, bids map[string]string, err error)
	GetById(ctx context.Context, orderId string) (*entity.Order, error)
	LoadScripts(ctx context.Context) error
}

type SessionRepository interface {
	GetSession(ctx context.Context, login string) (*entity.User, error)
	AddSession(ctx context.Context, user *entity.User) error
}

type Engine interface {
	AddOrder(ctx context.Context, order *entity.Order) error
	CancelOrder(ctx context.Context, orderId string) error
}
