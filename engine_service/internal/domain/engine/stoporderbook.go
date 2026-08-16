package engine

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"

type StopOrderBook struct {
	Book
}

func NewStopOrderBook(book Book) *StopOrderBook {
	return &StopOrderBook{
		Book: book,
	}
}

func (this *StopOrderBook) GetStopOrders(bestAskLevel, bestBidLevel int64) []*entity.Order {
	orders := make([]*entity.Order, 0)
	for bestBid := this.bid.GetFirst(); bestBid != nil && bestBidLevel >= bestBid.Price; bestBid = this.bid.GetFirst() {
		getOrder := bestBid.Orders.NewIterator()
		for stopOrder := getOrder.Next(); stopOrder != nil; stopOrder = getOrder.Next() {
			stopOrder.Content.ActivateStopOrder()

			orders = append(orders, stopOrder.Content)

			this.Remove(stopOrder.Content.Id)
		}
	}

	for bestAsk := this.ask.GetFirst(); bestAsk != nil && bestBidLevel <= bestAsk.Price; bestAsk = this.ask.GetFirst() {
		getOrder := bestAsk.Orders.NewIterator()
		for stopOrder := getOrder.Next(); stopOrder != nil; stopOrder = getOrder.Next() {
			stopOrder.Content.ActivateStopOrder()

			orders = append(orders, stopOrder.Content)

			this.Remove(stopOrder.Content.Id)
		}
	}

	return orders
}
