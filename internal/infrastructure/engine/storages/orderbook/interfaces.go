package orderbook

import enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"

type BookStorage interface {
	Add(order *enginetypes.EngineOrder)
	GetRange(quantity int64, price int64) []*enginetypes.PriceLevel
	GetFirst() *enginetypes.PriceLevel
	Delete(level int64)
}
