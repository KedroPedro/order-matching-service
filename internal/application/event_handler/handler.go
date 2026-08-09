package eventhandler

import (
	"context"
	"sync"
	"time"

	eventbatch "github.com/KedroPedro/order-matching-engine/internal/application/event_handler/event_batch"

	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/rs/zerolog/log"
)

type EventHandler struct {
	eventCh    <-chan *eventbatch.EventBatch
	marketRepo interfaces.MarketStateRepository
	orderRepo  interfaces.OrderRepository
	started    bool
}

const (
	eventProcessorsNumber = 4
)

func NewEventHandler(eventCh chan *eventbatch.EventBatch, marketRepo interfaces.MarketStateRepository, orderRepo interfaces.OrderRepository) *EventHandler {
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
		case batch, ok := <-this.eventCh:
			if !ok {
				return
			}

			events := batch.GetEvents()

			for i := range events {

				fCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				if err := this.marketRepo.ProcessEvent(fCtx, &events[i]); err != nil {
					cancel()
					log.Err(err).Stack().Send()
					continue
				}

				if err := this.orderRepo.ProcessEvent(fCtx, &events[i]); err != nil {
					cancel()
					log.Err(err).Stack().Send()
					continue
				}

				cancel()
			}
		case <-ctx.Done():
			return
		}
	}
}
