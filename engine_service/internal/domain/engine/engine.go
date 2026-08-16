package engine

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
)

const (
	ordersChanBuffer = 100_000
	cancelChanBuffer = 100_000
)

type Engine struct {
	orders         *OrderBook
	stopOrders     *StopOrderBook
	eventBatchChan chan<- *entity.EventBatch
	cancelChan     chan string
	dayChan        chan struct{}
	ordersChan     chan *entity.Order
	stopChan       chan struct{}
	closed         bool
}

func New(
	eventBatchChan chan<- *entity.EventBatch,
) *Engine {

	oBook := NewBook(
		NewSkipList(AscendingList),
		NewSkipList(DescendingList),
	)
	orders := NewOrderBook(oBook)

	sBook := NewBook(
		NewSkipList(AscendingList),
		NewSkipList(DescendingList),
	)
	stops := NewStopOrderBook(sBook)

	return &Engine{
		orders:         orders,
		stopOrders:     stops,
		eventBatchChan: eventBatchChan,
		ordersChan:     make(chan *entity.Order, ordersChanBuffer),
		dayChan:        make(chan struct{}),
		cancelChan:     make(chan string, cancelChanBuffer),
		stopChan:       make(chan struct{}),
		closed:         false,
	}
}

func (this *Engine) Process(ctx context.Context) {
	ordersClosed := false
	cancelsClosed := false
	dayClosed := false
	for {
		if ordersClosed && cancelsClosed && dayClosed {
			break
		}

		select {
		case incomingOrder, ok := <-this.ordersChan:
			if !ok {
				ordersClosed = true
				continue
			}

			this.processOrder(incomingOrder)

		case orderId, ok := <-this.cancelChan:
			if !ok {
				cancelsClosed = true
				continue
			}

			this.cancelOrder(orderId)

		case _, ok := <-this.dayChan:
			if !ok {
				dayClosed = true
				continue
			}

			this.closeDay()

		case <-this.stopChan:
			if this.closed {
				this.closed = false
			} else {
				this.closed = true
			}
		}
	}
}

func (this *Engine) EndDay() {
	this.dayChan <- struct{}{}
}

func (this *Engine) Stop() {
	if this.closed {
		return
	}

	this.stopChan <- struct{}{}
}

func (this *Engine) Start() {
	if !this.closed {
		return
	}

	this.stopChan <- struct{}{}
}

func (this *Engine) AddOrder(order *entity.Order) {
	this.ordersChan <- order
}

func (this *Engine) CancelOrder(id string) {
	this.cancelChan <- id
}

func (this *Engine) processOrder(order *entity.Order) {
	eventBatch := entity.NewEventBatch()

	if order.IsStopOrder() {
		this.stopOrders.Add(order)
		eventBatch.Add(order.SetPendingStatus())
		return
	}

	if this.closed && order.TimeInForce == entity.DAY {
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

	eventBatch := entity.NewEventBatch()

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
	eventBatch := entity.NewEventBatch()
	this.orders.Cancel(orderId, eventBatch)
	this.stopOrders.Cancel(orderId, eventBatch)

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}

}

func (this *Engine) closeDay() {
	eventBatch := entity.NewEventBatch()

	this.orders.CancelDayOrders(eventBatch)
	this.stopOrders.CancelDayOrders(eventBatch)

	if eventBatch.Len() > 0 {
		this.eventBatchChan <- eventBatch
	} else {
		eventBatch.Release()
	}
}
