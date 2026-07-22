package interfaces

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
)

type OrderRepository interface {
	AddToQueue(ctx context.Context, order *entity.Order) error
	ProcessEvent(ctx context.Context, event *entity.Event) error
}

type MarketStateRepository interface {
	AddToQueue(ctx context.Context, order *entity.Order) error
	ProcessEvent(ctx context.Context, event *entity.Event) error
	GetLastPrice(ctx context.Context, orderType entity.OrderType) (int64, error)
	GetState(ctx context.Context)
}

type Engine interface {
	AddToQueue(order *entity.Order) error
	Cancel(event *entity.Event) error
	Close()
	Open()
	IsClosed() bool
}
