package orderbook

import (
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type OrderBookStorage interface {
	Add(order *enginetypes.EngineOrder)
	GetRange(quantity int64) []*enginetypes.PriceLevel
	GetFirst() *enginetypes.PriceLevel
	Delete(level int64)
}

type OrderBook struct {
	ask       OrderBookStorage
	bid       OrderBookStorage
	orders    map[string]*enginetypes.EngineOrder
	dayOrders map[string]*enginetypes.EngineOrder
}

func New(askStorage, bidStorage OrderBookStorage) *OrderBook {
	return &OrderBook{
		ask:       askStorage,
		bid:       bidStorage,
		orders:    make(map[string]*enginetypes.EngineOrder),
		dayOrders: make(map[string]*enginetypes.EngineOrder),
	}
}

func (this *OrderBook) Add(order *enginetypes.EngineOrder) {
	if order == nil {
		return
	}

	switch order.GetType() {
	case enginetypes.Ask:
		this.ask.Add(order)
		this.orders[order.GetId()] = order
	case enginetypes.Bid:
		this.bid.Add(order)
		this.orders[order.GetId()] = order
	default:
		return
	}

	if order.GetTimeInForce() == enginetypes.DAY {
		this.dayOrders[order.GetId()] = order
	}
}

func (this *OrderBook) Remove(orderId string) {
	if order, ok := this.orders[orderId]; ok {
		order.Delete()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}

}

func (this *OrderBook) CancelDayOrders() {
	for key, value := range this.dayOrders {
		value.SetExpiredStatus()
		this.Remove(key)

		delete(this.orders, key)
		delete(this.dayOrders, key)
	}
}

func (this *OrderBook) Cancel(orderId string) {
	if order, ok := this.orders[orderId]; ok {
		order.Delete()
		order.SetCanceledStatus()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}
}

func (this *OrderBook) GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder {
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

func (this *OrderBook) Match(order *enginetypes.EngineOrder) {
	var levels []*enginetypes.PriceLevel

	switch order.GetType() {
	case enginetypes.Ask:
		levels = this.ask.GetRange(order.GetQuantity())
	case enginetypes.Bid:
		levels = this.bid.GetRange(order.GetQuantity())
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
		level.DecreaseQuantity(currUsedQuantity)

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
			if order.GetType() == enginetypes.Ask {
				this.ask.Add(order)
			} else {
				this.bid.Add(order)
			}
			this.dayOrders[order.GetId()] = order
		}

	case enginetypes.GTC:
		if order.GetStatus() != enginetypes.Filled {
			if order.GetType() == enginetypes.Ask {
				this.ask.Add(order)
			} else {
				this.bid.Add(order)
			}
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
