package interfaces

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
)

type UserRepository interface {
	ReserveBalance(ctx context.Context, userId string, toReserve int64) error
	ReleaseBalance(ctx context.Context, userId string, toReserve int64) error
	UpdateBalance(ctx context.Context, userId string, reserved int64, spent int64) error
}

type OrderRepository interface {
	AddToQueue(ctx context.Context, order *entity.Order) error
	ProcessEvent(ctx context.Context, event *entity.Event) error
}

type MarketStateRepository interface {
	AddToQueue(ctx context.Context, order *entity.Order) error
	ProcessEvent(ctx context.Context, event *entity.Event) error
	GetBestPrice(ctx context.Context, orderType entity.OrderType) (int64, error)
	GetState(ctx context.Context)
}

type Engine interface {
	AddToQueue(order *entity.Order) error
	Cancel(event *entity.Event) error
	Close()
	Open()
	IsClosed() bool
}
