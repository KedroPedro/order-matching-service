package interfaces

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
)

type OrderRepository interface {
	ProcessEvent(ctx context.Context, event *entity.Event) error
}

type MarketStateRepository interface {
	ProcessEvent(ctx context.Context, event *entity.Event) error
	GetBestPrice(ctx context.Context, orderType entity.OrderType) (int64, error)
	GetById(ctx context.Context, orderId string) (*entity.Order, error)
	LoadScripts(ctx context.Context) error
}
