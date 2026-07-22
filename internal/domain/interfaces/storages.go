package interfaces

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
)

type OrderRepository interface {
	AddToQueue(ctx context.Context, order *entity.Order) error
	ProcessEvent(ctx context.Context, event *entity.Event) error
}
