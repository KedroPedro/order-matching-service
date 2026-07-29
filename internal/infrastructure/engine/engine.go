package engine

import (
	"context"
	"fmt"
	"runtime"
	"time"

	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type StopOrderBook interface {
	Add(order *enginetypes.EngineOrder)
	GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder
	Cancel(orderId string)
	CancelDayOrders()
}

type OrderBook interface {
	BestAskPrice() int64
	BestBidPrice() int64
	Match(order *enginetypes.EngineOrder)
	Remove(orderId string)
	Cancel(orderId string)
	CancelDayOrders()
}

type OrderCollection interface {
	GetOrders() func() *enginetypes.EngineOrder
	GetLevel() int64
	GetQuantity() int64
	DecreaseQuantity(int64)
}

type Engine struct {
	orders     OrderBook
	stopOrders StopOrderBook
	closed     bool
}

func New(
	ctx context.Context,
	orders OrderBook,
	stops StopOrderBook,
	orderChan <-chan *enginetypes.EngineOrder,
	cancelChan <-chan string,
	endChan chan<- struct{},
	dayChan <-chan struct{},
) *Engine {
	engine := &Engine{
		orders:     orders,
		stopOrders: stops,
		closed:     false,
	}

	go func() {
		ordersClosed := false
		cancelsClosed := false
		dayClosed := false
		for {
			if ordersClosed && cancelsClosed && dayClosed {
				endChan <- struct{}{}
				return
			}

			select {
			case incomingOrder, ok := <-orderChan:
				if !ok {
					ordersClosed = true
					continue
				}

				incomingOrder.SetNewStatus()

				engine.processOrder(incomingOrder)

			case orderId, ok := <-cancelChan:
				if !ok {
					cancelsClosed = true
					continue
				}

				engine.cancelOrder(orderId)

			case _, ok := <-dayChan:
				if !ok {
					dayClosed = true
					continue
				}

				if engine.closed {
					engine.closed = false
				} else {
					engine.closed = true
					engine.close()
				}
			}
		}
	}()

	return engine
}

func (this *Engine) processOrder(order *enginetypes.EngineOrder) {
	if order.IsStopOrder() {
		this.stopOrders.Add(order)
		order.SetPendingStatus()
		return
	}

	if this.closed && order.GetTimeInForce() == enginetypes.DAY {
		order.SetExpiredStatus()
		return
	}

	this.orders.Match(order)

	this.processStopOrders()
}

func (this *Engine) processStopOrders() {
	orders := this.stopOrders.GetStopOrders(this.orders.BestAskPrice(), this.orders.BestBidPrice())

	for _, order := range orders {
		this.orders.Match(order)
	}
}

func (this *Engine) cancelOrder(orderId string) {
	this.orders.Cancel(orderId)
	this.stopOrders.Cancel(orderId)
}

func (this *Engine) close() {
	this.orders.CancelDayOrders()
	this.stopOrders.CancelDayOrders()
}
