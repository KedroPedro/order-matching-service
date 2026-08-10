package starter

import (
	"context"

	eventbatch "github.com/KedroPedro/order-matching-engine/internal/application/event_handler/event_batch"
	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/orderbook"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
)

const (
	engineOrderChanBuffer      = 100000
	engineCancelChanBuffer     = 100000
	engineEventBatchChanBuffer = 10000
)

type Starter struct {
	engine           *engine.Engine
	orderChan        chan *entity.Order
	dayChan          chan struct{}
	cancelChan       chan string
	engineEndChan    chan struct{}
	eventBatchChan   chan *eventbatch.EventBatch
	engineOrderChan  chan *enginetypes.EngineOrder
	engineCancelChan chan string
	engineDayChan    chan struct{}
	closed           bool
}

func New(
	ctx context.Context,
) (*Starter, chan *eventbatch.EventBatch) {

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

	eventBatchChan := make(chan *eventbatch.EventBatch, engineEventBatchChanBuffer)
	engineOrderChan := make(chan *enginetypes.EngineOrder, engineOrderChanBuffer)
	engineEndChan := make(chan struct{})
	engineCancelChan := make(chan string, engineCancelChanBuffer)
	engineDayChan := make(chan struct{})

	newEngine := engine.New(
		ctx,
		orderBook,
		stopBook,
		eventBatchChan,
		engineOrderChan,
		engineCancelChan,
		engineEndChan,
		engineDayChan,
	)

	newStarter := &Starter{
		engine:           newEngine,
		orderChan:        make(chan *entity.Order, engineCancelChanBuffer),
		dayChan:          make(chan struct{}),
		cancelChan:       make(chan string, engineCancelChanBuffer),
		engineEndChan:    engineEndChan,
		engineOrderChan:  engineOrderChan,
		engineCancelChan: engineCancelChan,
		engineDayChan:    engineDayChan,
		eventBatchChan:   eventBatchChan,
		closed:           false,
	}

	return newStarter, eventBatchChan
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

			newEngineOrder := enginetypes.NewEngineOrder(order, nil)

			newEngineOrder.SetPendingStatus()

			this.engineOrderChan <- newEngineOrder

		case orderId, ok := <-this.cancelChan:
			if !ok {
				<-this.engineEndChan
				return
			}

			this.engineCancelChan <- orderId

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

func (this *Starter) Cancel(orderId string) {
	this.cancelChan <- orderId
}
