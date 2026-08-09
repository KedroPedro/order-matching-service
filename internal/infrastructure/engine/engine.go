package engine

import (
	"context"

	eventbatch "github.com/KedroPedro/order-matching-engine/internal/application/event_handler/event_batch"

	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type Book interface {
	Cancel(orderId string, eventBatch *eventbatch.EventBatch)
	CancelDayOrders(eventBatch *eventbatch.EventBatch)
}

type StopOrderBook interface {
	Book
	Add(order *enginetypes.EngineOrder)
	GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder
}

type OrderBook interface {
	Book
	BestAskPrice() int64
	BestBidPrice() int64
	Match(order *enginetypes.EngineOrder, eventBatch *eventbatch.EventBatch)
	Remove(orderId string)
}

type OrderCollection interface {
	GetOrders() func() *enginetypes.EngineOrder
	GetLevel() int64
	GetQuantity() int64
	DecreaseQuantity(int64)
}

type Engine struct {
	orders         OrderBook
	stopOrders     StopOrderBook
	eventBatchChan chan<- *eventbatch.EventBatch
	closed         bool
}

func New(
	ctx context.Context,
	orders OrderBook,
	stops StopOrderBook,
	eventBatchChan chan<- *eventbatch.EventBatch,
	orderChan <-chan *enginetypes.EngineOrder,
	cancelChan <-chan string,
	endChan chan<- struct{},
	dayChan <-chan struct{},
) *Engine {
	engine := &Engine{
		orders:         orders,
		stopOrders:     stops,
		closed:         false,
		eventBatchChan: eventBatchChan,
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
	eventBatch := eventbatch.New()

	if order.IsStopOrder() {
		this.stopOrders.Add(order)
		eventBatch.Add(order.SetPendingStatus())
		return
	}

	if this.closed && order.GetTimeInForce() == enginetypes.DAY {
		eventBatch.Add(order.SetExpiredStatus())
		return
	}

	eventBatch.Add(order.SetNewStatus())

	this.orders.Match(order, eventBatch)

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}

	this.processStopOrders()
}

func (this *Engine) processStopOrders() {

	orders := this.stopOrders.GetStopOrders(this.orders.BestAskPrice(), this.orders.BestBidPrice())

	eventBatch := eventbatch.New()

	for _, order := range orders {
		eventBatch.Add(order.SetNewStatus())
		this.orders.Match(order, eventBatch)
	}

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}
}

func (this *Engine) cancelOrder(orderId string) {
	eventBatch := eventbatch.New()
	this.orders.Cancel(orderId, eventBatch)
	this.stopOrders.Cancel(orderId, eventBatch)

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}

}

func (this *Engine) close() {
	eventBatch := eventbatch.New()

	this.orders.CancelDayOrders(eventBatch)
	this.stopOrders.CancelDayOrders(eventBatch)

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}

}
