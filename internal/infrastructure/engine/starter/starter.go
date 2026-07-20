package starter

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/orderbook"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
)

const (
	engineOrderChanBuffer  = 1000
	engineCancelChanBuffer = 1000
)

type Starter struct {
	engine          *engine.Engine
	orderChan       <-chan *entity.Order
	engineOrderChan chan<- *enginetypes.EngineOrder
	closed          bool
}

func New(
	ctx context.Context,
	orderChan <-chan *entity.Order,
	cancelChan <-chan entity.Event,
	eventChan chan entity.Event, dayChan <-chan struct{},
) *Starter {
	orderBook := orderbook.New(
		skiplist.New(skiplist.AscendingList),
		skiplist.New(skiplist.DescendingList),
	)

	stopBook := orderbook.New(
		skiplist.New(skiplist.AscendingList),
		skiplist.New(skiplist.DescendingList),
	)

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
		engine:          newEngine,
		engineOrderChan: engineOrderChan,
		closed:          false,
	}

	go func() {
		for {
			select {
			case order, ok := <-orderChan:
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

			case cancelEvent, ok := <-cancelChan:
				if !ok {
					<-engineEndChan
					return
				}

				if cancelEvent.EventType != entity.OrderCancelled {
					continue
				}

				engineCancelChan <- cancelEvent.OrderId

			case <-dayChan:
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

	return newStarter
}
