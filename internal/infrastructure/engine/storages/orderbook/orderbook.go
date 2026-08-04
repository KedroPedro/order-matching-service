package orderbook

import (
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type OrderBook struct {
	Book
}

func NewOrderBook(book Book) *OrderBook {
	return &OrderBook{
		Book: book,
	}
}

func (this *OrderBook) Match(order *enginetypes.EngineOrder) {
	var levels []*enginetypes.PriceLevel

	switch order.GetType() {
	case enginetypes.Ask:
		levels = this.bid.GetRange(order.GetQuantity())
	case enginetypes.Bid:
		levels = this.ask.GetRange(order.GetQuantity())
	default:
		return
	}

	var totalPrice int64 = 0
	var totalUsedQuantity int64 = 0

	for _, level := range levels {
		unf, rest := order.Fill(level.GetQuantity(), level.GetLevel())

		totalPrice += (level.GetQuantity() - rest) * level.GetLevel()

		currUsedQuantity := level.GetQuantity() - rest

		totalUsedQuantity += currUsedQuantity

		if unf == 0 {
			break
		}

		if order.GetReserve() < level.GetLevel() {
			break
		}
	}

	switch order.GetTimeInForce() {
	case enginetypes.DAY:
		if order.GetStatus() != enginetypes.Filled {
			this.Add(order)
			this.dayOrders[order.GetId()] = order
		}

	case enginetypes.GTC:
		if order.GetStatus() != enginetypes.Filled {
			this.Add(order)
		}

	case enginetypes.FOK:
		if order.GetStatus() != enginetypes.Filled {
			order.SetRejectedStatus()
		}
		return

	case enginetypes.IOC:
		if order.GetStatus() != enginetypes.Filled {
			order.SetCanceledStatus()
		}
	}

outer:
	for _, level := range levels {
		getFilledOrder := level.GetOrders()
		for filledOrder := getFilledOrder(); filledOrder != nil && totalUsedQuantity != 0; filledOrder = getFilledOrder() {
			_, rest := filledOrder.Fill(totalUsedQuantity, filledOrder.GetLevel())

			level.DecreaseQuantity(totalUsedQuantity - rest)

			totalUsedQuantity -= totalUsedQuantity - rest

			if filledOrder.GetStatus() == enginetypes.Filled {
				this.Remove(filledOrder.GetId())
			}

			if totalUsedQuantity == 0 {
				break outer
			}
		}
	}
}

func (this *OrderBook) BestBidPrice() int64 {
	order := this.bid.GetFirst()
	if order != nil {
		return order.GetLevel()
	}
	return 0
}

func (this *OrderBook) BestAskPrice() int64 {
	order := this.ask.GetFirst()
	if order != nil {
		return order.GetLevel()
	}
	return 0

}
