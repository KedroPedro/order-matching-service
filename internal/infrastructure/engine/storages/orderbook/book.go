package orderbook

import enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"

type Book struct {
	ask       BookStorage
	bid       BookStorage
	orders    map[string]*enginetypes.EngineOrder
	dayOrders map[string]*enginetypes.EngineOrder
}

func NewBook(askStorage, bidStorage BookStorage) Book {
	return Book{
		ask:       askStorage,
		bid:       bidStorage,
		orders:    make(map[string]*enginetypes.EngineOrder, 1024),
		dayOrders: make(map[string]*enginetypes.EngineOrder, 1024),
	}
}

func (this *Book) Add(order *enginetypes.EngineOrder) {
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

func (this *Book) Remove(orderId string) {
	if order, ok := this.orders[orderId]; ok {
		order.Delete()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}
}

func (this *Book) CancelDayOrders() {
	for key, value := range this.dayOrders {
		value.SetExpiredStatus()
		value.Delete()
		delete(this.orders, key)
		delete(this.dayOrders, key)
	}
}

func (this *Book) Cancel(orderId string) {
	if order, ok := this.orders[orderId]; ok {
		order.SetCanceledStatus()
		order.Delete()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}
}
