package eventhandler

import (
	"context"
	"sync"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/rs/zerolog/log"
)

type EventHandler struct {
	eventCh    <-chan *entity.Event
	marketRepo interfaces.MarketStateRepository
	orderRepo  interfaces.OrderRepository
	started    bool
}

const (
	eventProcessorsNumber = 1 //when more than 1 can cause a race conditions
)

func NewEventHandler(eventCh chan *entity.Event, marketRepo interfaces.MarketStateRepository, orderRepo interfaces.OrderRepository) *EventHandler {
	return &EventHandler{
		eventCh:    eventCh,
		marketRepo: marketRepo,
		orderRepo:  orderRepo,
	}
}

func (this *EventHandler) Start(ctx context.Context) {
	if !this.started {
		this.started = true
	} else {
		return
	}

	wg := &sync.WaitGroup{}
	for range eventProcessorsNumber {
		wg.Add(1)
		go func() {
			this.processEvent(ctx)
			wg.Done()
		}()
	}
	wg.Wait()

}

func (this *EventHandler) processEvent(ctx context.Context) {
	for {
		select {
		case event, ok := <-this.eventCh:
			if !ok {
				return
			}

			fCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
			if err := this.marketRepo.ProcessEvent(fCtx, event); err != nil {
				cancel()
				log.Err(err).Stack().Send()
				continue
			}

			if err := this.orderRepo.ProcessEvent(fCtx, event); err != nil {
				cancel()
				log.Err(err).Stack().Send()
				continue
			}

			cancel()

		case <-ctx.Done():
			return
		}
	}
}
