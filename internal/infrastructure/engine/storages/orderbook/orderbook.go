package orderbook

import (
	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
)

type OrderBook struct {
	ask    *skiplist.SkipList
	bid    *skiplist.SkipList
	stop   *skiplist.SkipList
	orders map[string]entity.Order
}

func (this *OrderBook) ProccessOrder(order *entity.Order) {

}

func (this *OrderBook) DeleteOrder(orderId string) {

}

func (this *OrderBook) addAskOrder(order *entity.Order) {

}

func (this *OrderBook) addBidOrder(order *entity.Order) {

}

func (this *OrderBook) addStopOrder(order *entity.Order) {

}

func (this *OrderBook) proccessMarketOrder(order *entity.Order) {

}
