package starter

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/orderbook"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
)

const (
	engineOrderChanBuffer  = 100000
	engineCancelChanBuffer = 100000
	engineEventChanBuffer  = 100000
)

type Starter struct {
	engine           *engine.Engine
	orderChan        chan *entity.Order
	dayChan          chan struct{}
	cancelChan       chan *entity.Event
	engineEndChan    chan struct{}
	eventChan        chan *entity.Event
	engineOrderChan  chan *enginetypes.EngineOrder
	engineCancelChan chan string
	engineDayChan    chan struct{}
	closed           bool
}

func New(
	ctx context.Context,
) (*Starter, chan *entity.Event) {

	orderBook := orderbook.NewOrderBook(
		orderbook.NewBook(
			skiplist.New(skiplist.AscendingList),
			skiplist.New(skiplist.DescendingList),
		),
	)

	stopBook := orderbook.NewStopOrderBook(
		orderbook.NewBook(
			skiplist.New(skiplist.AscendingList),
			skiplist.New(skiplist.DescendingList),
		),
	)

	eventChan := make(chan *entity.Event, engineEventChanBuffer)
	engineOrderChan := make(chan *enginetypes.EngineOrder, engineOrderChanBuffer)
	engineEndChan := make(chan struct{})
	engineCancelChan := make(chan string, engineCancelChanBuffer)
	engineDayChan := make(chan struct{})

	newEngine := engine.New(
		ctx,
		orderBook,
		stopBook,
		engineOrderChan,
		engineCancelChan,
		engineEndChan,
		engineDayChan,
	)

	newStarter := &Starter{
		engine:           newEngine,
		orderChan:        make(chan *entity.Order, engineCancelChanBuffer),
		dayChan:          make(chan struct{}),
		cancelChan:       make(chan *entity.Event, engineCancelChanBuffer),
		engineEndChan:    engineEndChan,
		engineOrderChan:  engineOrderChan,
		engineCancelChan: engineCancelChan,
		engineDayChan:    engineDayChan,
		eventChan:        eventChan,
		closed:           false,
	}

	return newStarter, eventChan
}

func (this *Starter) Start(ctx context.Context) {
	for {
		select {
		case order, ok := <-this.orderChan:
			if !ok {
				<-this.engineEndChan
				return
			}

			if this.closed {
				continue
			}

			newEngineOrder := enginetypes.NewEngineOrder(order, this.eventChan)

			newEngineOrder.SetPendingStatus()

			this.engineOrderChan <- newEngineOrder

		case cancelEvent, ok := <-this.cancelChan:
			if !ok {
				<-this.engineEndChan
				return
			}

			if cancelEvent.GetType() != entity.OrderCancelled {
				continue
			}

			payload, ok := cancelEvent.GetPayload().(entity.OrderRemovalPayload)
			if !ok {
				continue
			}

			this.engineCancelChan <- payload.Order.Id

		case <-this.dayChan:
			if this.closed {
				this.closed = false
			} else {
				this.closed = true
			}
			this.engineDayChan <- struct{}{}

		case <-ctx.Done():
			<-this.engineEndChan
			return

		}
	}
}

func (this Starter) IsClosed() bool {
	return this.closed
}

func (this *Starter) Open() {
	if this.closed {
		this.dayChan <- struct{}{}
	}
}

func (this *Starter) Close() {
	if !this.closed {
		this.dayChan <- struct{}{}
	}
}

func (this *Starter) AddToQueue(order *entity.Order) error {
	if order == nil {
		return errs.NewEngineError("nil order")
	}

	this.orderChan <- order
	return nil
}

func (this *Starter) Cancel(event *entity.Event) error {
	if event == nil {
		return errs.NewEngineError("nil event")
	}

	this.cancelChan <- event
	return nil
}
