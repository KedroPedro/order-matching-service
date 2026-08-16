package engine

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"

type OrderBook struct {
	Book
}

func NewOrderBook(book Book) *OrderBook {
	return &OrderBook{
		Book: book,
	}
}

func (this *OrderBook) Match(order *entity.Order, eventBatch *entity.EventBatch) {
	var levels []*PriceLevel

	switch order.Type {
	case entity.Ask:
		levels = this.bid.GetRange(order.GetUnfilledQuantity(), order.Price)
	case entity.Bid:
		levels = this.ask.GetRange(order.GetUnfilledQuantity(), order.Price)
	default:
		return
	}

	if len(levels) == 0 {
		switch order.TimeInForce {
		case entity.GTC:
			this.Add(order)
			return
		case entity.DAY:
			this.Add(order)
			return
		case entity.IOC:
			eventBatch.Add(order.SetCanceledStatus())
			return
		case entity.FOK:
			eventBatch.Add(order.SetRejectedStatus())
			return
		}
	}

	if order.TimeInForce == entity.FOK {
		totalQuantity := int64(0)
		for i := 0; i < len(levels); i++ {
			totalQuantity += levels[i].TotalQuantity
		}

		if totalQuantity < order.GetUnfilledQuantity() {
			eventBatch.Add(order.SetRejectedStatus())
			return
		}
	}

	for i := 0; i < len(levels) && order.Status != entity.Filled; i++ {
		if levels[i].Price > order.Reserve {
			break
		}

		var usedQuantity int64

		if order.GetUnfilledQuantity() < levels[i].TotalQuantity {
			usedQuantity = order.GetUnfilledQuantity()
			eventBatch.Add(order.Fill(usedQuantity, levels[i].Price))
		} else {
			usedQuantity = levels[i].TotalQuantity
			eventBatch.Add(order.Fill(usedQuantity, levels[i].Price))
		}

		if order.GetUnfilledQuantity() == 0 {
			eventBatch.Add(order.SetFilledStatus())
		}

		orderIter := levels[i].Orders.NewIterator()

		levels[i].TotalQuantity -= usedQuantity

		for oppositeOrder := orderIter.Next(); oppositeOrder != nil && usedQuantity > 0; oppositeOrder = orderIter.Next() {
			if oppositeOrder.Content.GetUnfilledQuantity() < usedQuantity {
				usedQuantity -= oppositeOrder.Content.GetUnfilledQuantity()
				eventBatch.Add(oppositeOrder.Content.Fill(oppositeOrder.Content.GetUnfilledQuantity(), levels[i].Price))
			} else {
				eventBatch.Add(oppositeOrder.Content.Fill(usedQuantity, levels[i].Price))
				usedQuantity = 0
			}

			if oppositeOrder.Content.GetUnfilledQuantity() == 0 {
				eventBatch.Add(oppositeOrder.Content.SetFilledStatus())
				this.Remove(oppositeOrder.Content.Id)
			}
		}
	}

	if order.Status == entity.Filled {
		return
	}

	if order.TimeInForce == entity.IOC && order.Status != entity.Filled {
		eventBatch.Add(order.SetCanceledStatus())
		return
	}

	if order.TimeInForce == entity.FOK {
		eventBatch.Add(order.SetRejectedStatus())
		return
	}

	this.Add(order)
}

func (this *OrderBook) BestBidPrice() int64 {
	order := this.bid.GetFirst()
	if order != nil {
		return order.Price
	}
	return 0
}

func (this *OrderBook) BestAskPrice() int64 {
	order := this.ask.GetFirst()
	if order != nil {
		return order.Price
	}
	return 0

}
