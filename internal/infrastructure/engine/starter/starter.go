package starter

import (
	"context"
	"errors"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/orderbook"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
)

const (
	engineOrderChanBuffer  = 1000
	engineCancelChanBuffer = 1000
	engineEventChanBuffer  = 1000
)

type Starter struct {
	engine     *engine.Engine
	orderChan  chan *entity.Order
	dayChan    chan struct{}
	cancelChan chan *entity.Event
	closed     bool
}

func New(
	ctx context.Context,
) (*Starter, chan *entity.Event) {
	orderBook := orderbook.New(
		skiplist.New(skiplist.AscendingList),
		skiplist.New(skiplist.DescendingList),
	)

	stopBook := orderbook.New(
		skiplist.New(skiplist.AscendingList),
		skiplist.New(skiplist.DescendingList),
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
		engine:     newEngine,
		orderChan:  make(chan *entity.Order, engineCancelChanBuffer),
		dayChan:    make(chan struct{}),
		cancelChan: make(chan *entity.Event, engineCancelChanBuffer),
		closed:     false,
	}

	go func() {
		for {
			select {
			case order, ok := <-newStarter.orderChan:
				if !ok {
					<-engineEndChan
					return
				}

				if newStarter.closed {
					continue
				}

				newEngineOrder := enginetypes.NewEngineOrder(order, eventChan)

				newEngineOrder.SetPendingStatus()

				engineOrderChan <- newEngineOrder

			case cancelEvent, ok := <-newStarter.cancelChan:
				if !ok {
					<-engineEndChan
					return
				}

				if cancelEvent.EventType != entity.OrderCancelled {
					continue
				}

				engineCancelChan <- cancelEvent.OrderId

			case <-newStarter.dayChan:
				if newStarter.closed {
					newStarter.closed = false
				} else {
					newStarter.closed = true
				}
				engineDayChan <- struct{}{}

			case <-ctx.Done():
				<-engineEndChan
				return

			}
		}
	}()

	return newStarter, eventChan
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
		return errors.New("nil order")
	}

	this.orderChan <- order
	return nil
}

func (this *Starter) Cancel(event *entity.Event) error {
	if event == nil {
		return errors.New("nil event")
	}

	this.cancelChan <- event
	return nil
}
