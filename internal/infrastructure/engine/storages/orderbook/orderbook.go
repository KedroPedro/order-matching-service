package orderbook

import (
	eventbatch "github.com/KedroPedro/order-matching-engine/internal/application/event_handler/event_batch"
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

func (this *OrderBook) Match(order *enginetypes.EngineOrder, eventBatch *eventbatch.EventBatch) {
	var levels []*enginetypes.PriceLevel

	switch order.GetType() {
	case enginetypes.Ask:
		levels = this.bid.GetRange(order.GetUnfilledQuantity(), order.GetLevel())
	case enginetypes.Bid:
		levels = this.ask.GetRange(order.GetUnfilledQuantity(), order.GetLevel())
	default:
		return
	}

	if len(levels) == 0 {
		switch order.GetTimeInForce() {
		case enginetypes.GTC:
			this.Add(order)
			return
		case enginetypes.DAY:
			this.Add(order)
			this.dayOrders[order.GetId()] = order
			return
		case enginetypes.IOC:
			eventBatch.Add(order.SetCanceledStatus())
			return
		case enginetypes.FOK:
			eventBatch.Add(order.SetRejectedStatus())
			return
		}
	}

	if order.GetTimeInForce() == enginetypes.FOK {
		totalQuantity := int64(0)
		for i := 0; i < len(levels); i++ {
			totalQuantity += levels[i].GetQuantity()
		}

		if totalQuantity < order.GetUnfilledQuantity() {
			eventBatch.Add(order.SetRejectedStatus())
			return
		}
	}

	for i := 0; i < len(levels) && order.GetStatus() != enginetypes.Filled; i++ {
		if levels[i].GetLevel() > order.GetReserve() {
			break
		}

		var usedQuantity int64

		if order.GetUnfilledQuantity() < levels[i].GetQuantity() {
			usedQuantity = order.GetUnfilledQuantity()
			eventBatch.Add(order.Fill(usedQuantity, levels[i].GetLevel()))
		} else {
			usedQuantity = levels[i].GetQuantity()
			eventBatch.Add(order.Fill(usedQuantity, levels[i].GetLevel()))
		}

		if order.GetUnfilledQuantity() == 0 {
			eventBatch.Add(order.SetFilledStatus())
		}

		orderIter := levels[i].GetOrdersIterator()

		levels[i].DecreaseQuantity(usedQuantity)

		for filledOrder := orderIter.Next(); filledOrder != nil && usedQuantity > 0; filledOrder = orderIter.Next() {
			if filledOrder.GetUnfilledQuantity() < usedQuantity {
				usedQuantity -= filledOrder.GetUnfilledQuantity()
				eventBatch.Add(filledOrder.Fill(filledOrder.GetUnfilledQuantity(), levels[i].GetLevel()))
			} else {
				eventBatch.Add(filledOrder.Fill(usedQuantity, levels[i].GetLevel()))
				usedQuantity = 0
			}

			if filledOrder.GetUnfilledQuantity() == 0 {
				eventBatch.Add(filledOrder.SetFilledStatus())
				this.Remove(filledOrder.GetId())
			}
		}
	}

	if order.GetStatus() == enginetypes.Filled {
		return
	}

	if order.GetTimeInForce() == enginetypes.IOC && order.GetStatus() != enginetypes.Filled {
		eventBatch.Add(order.SetCanceledStatus())
		return
	}

	if order.GetTimeInForce() == enginetypes.FOK {
		eventBatch.Add(order.SetRejectedStatus())
		return
	}

	this.Add(order)

	if order.GetTimeInForce() == enginetypes.DAY {
		this.dayOrders[order.GetId()] = order
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
