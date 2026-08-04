package orderbook

import enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"

type StopOrderBook struct {
	Book
}

func NewStopOrderBook(book Book) *StopOrderBook {
	return &StopOrderBook{
		Book: book,
	}

}

func (this *StopOrderBook) GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder {
	orders := make([]*enginetypes.EngineOrder, 0)
	for bestBid := this.bid.GetFirst(); bestBid != nil && bestBidLevel >= bestBid.GetLevel(); bestBid = this.bid.GetFirst() {
		getOrder := bestBid.GetOrders()
		for stopOrder := getOrder(); stopOrder != nil; stopOrder = getOrder() {
			stopOrder.ActivateStopOrder()

			orders = append(orders, stopOrder)

			this.Remove(stopOrder.GetId())
		}
	}

	for bestAsk := this.ask.GetFirst(); bestAsk != nil && bestBidLevel <= bestAsk.GetLevel(); bestAsk = this.ask.GetFirst() {
		getOrder := bestAsk.GetOrders()
		for stopOrder := getOrder(); stopOrder != nil; stopOrder = getOrder() {
			stopOrder.ActivateStopOrder()

			orders = append(orders, stopOrder)

			this.Remove(stopOrder.GetId())
		}
	}

	return orders
}
