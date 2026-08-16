package engine

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"

type Book struct {
	ask       *SkipList
	bid       *SkipList
	orders    map[string]*EngineOrder
	dayOrders map[string]*EngineOrder
}

func NewBook(askStorage, bidStorage *SkipList) Book {
	return Book{
		ask:       askStorage,
		bid:       bidStorage,
		orders:    make(map[string]*EngineOrder, 1024),
		dayOrders: make(map[string]*EngineOrder, 1024),
	}
}

func (this *Book) Add(order *entity.Order) {
	if order == nil {
		return
	}

	engineOrder := NewEngineOrder(order)

	switch order.Type {
	case entity.Ask:
		this.ask.Add(engineOrder)
		this.orders[order.Id] = engineOrder
	case entity.Bid:
		this.bid.Add(engineOrder)
		this.orders[order.Id] = engineOrder
	default:
		return
	}

	if order.TimeInForce == entity.DAY {
		this.dayOrders[order.Id] = engineOrder
	}
}

func (this *Book) Remove(orderId string) {
	if order, ok := this.orders[orderId]; ok {
		order.Delete()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}
}

func (this *Book) CancelDayOrders(eventBatch *entity.EventBatch) {
	for key, value := range this.dayOrders {
		eventBatch.Add(value.Content.SetExpiredStatus())
		value.Delete()
		delete(this.orders, key)
		delete(this.dayOrders, key)
	}
}

func (this *Book) Cancel(orderId string, eventBatch *entity.EventBatch) {
	if order, ok := this.orders[orderId]; ok {
		eventBatch.Add(order.Content.SetCanceledStatus())
		order.Delete()
		delete(this.orders, orderId)
		delete(this.dayOrders, orderId)
	}
}
