package orderbook

import (
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/doublylinkedlist"
)

type OrderBookStorage interface {
	Add(order *enginetypes.EngineOrder)
	GetRange(size int64) []*doublylinkedlist.DoublyLinkedList
	GetFirst() *doublylinkedlist.DoublyLinkedList
	Get(level int64) *doublylinkedlist.DoublyLinkedList
	Delete(level int64)
	DeleteFirst()
}

type OrderBook struct {
	ask    OrderBookStorage
	bid    OrderBookStorage
	orders map[string]*enginetypes.EngineOrder
}

func (this *OrderBook) Add(order *enginetypes.EngineOrder) error {
	switch order.GetType() {
	case enginetypes.Ask:
		this.ask.Add(order)
		this.orders[order.GetId()] = order
	case enginetypes.Bid:
		this.bid.Add(order)
		this.orders[order.GetId()] = order
	default: //TODO: add error
		return nil
	}
	return nil
}

func (this *OrderBook) Remove(orderId string) {
	order, ok := this.orders[orderId]

	if ok {
		order.Delete()
	}
}

func (this *OrderBook) Match(order *enginetypes.EngineOrder) (used []*enginetypes.EngineOrder) {
	lists := this.ask.GetRange(order.GetQuantity())
	var totalQuantity int64 = 0

	for _, list := range lists {
		for currNode := list.GetFirst(); currNode != nil && totalQuantity < order.GetQuantity(); currNode = currNode.Next {
			totalQuantity += currNode.Value.GetQuantity()
			used = append(used, currNode.Value)
		}
	}

	return used
}

func (this *OrderBook) BestBid() *enginetypes.EngineOrder {
	return this.bid.GetFirst().GetFirst().Value
}

func (this *OrderBook) BestAsk() *enginetypes.EngineOrder {
	return this.bid.GetFirst().GetFirst().Value
}
