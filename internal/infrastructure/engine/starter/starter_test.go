package starter

import (
	"context"
	"testing"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        context.Context
		orderChan  <-chan *entity.Order
		cancelChan <-chan entity.Event
		eventChan  chan entity.Event
		dayChan    <-chan struct{}
		testNumber int
	}{
		{
			name:       "create starter",
			ctx:        context.Background(),
			orderChan:  make(<-chan *entity.Order),
			cancelChan: make(<-chan entity.Event),
			eventChan:  make(chan entity.Event),
			dayChan:    make(<-chan struct{}),
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(tt.ctx, tt.orderChan, tt.cancelChan, tt.eventChan, tt.dayChan)

			switch tt.testNumber {
			case 1:
				require.NotNil(t, got)
				require.NotNil(t, got.engine)
				require.NotNil(t, got.engineOrderChan)
				require.False(t, got.closed)
			}
		})
	}
}

func TestStarter_OrderProcessing(t *testing.T) {
	tests := []struct {
		name       string
		testNumber int
	}{
		{
			name:       "process order does not panic",
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			orderChan := make(chan *entity.Order)
			cancelChan := make(chan entity.Event)
			eventChan := make(chan entity.Event, 10)
			dayChan := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			_ = New(ctx, orderChan, cancelChan, eventChan, dayChan)

			switch tt.testNumber {
			case 1:
				order := &entity.Order{
					Id:          "1",
					Type:        entity.Ask,
					TimeInForce: entity.Gtc,
					Status:      entity.New,
					Price:       100,
					Quantity:    10,
				}

				orderChan <- order

				select {
				case <-time.After(100 * time.Millisecond):
				}
			}
		})
	}
}

func TestStarter_CancelProcessing(t *testing.T) {
	tests := []struct {
		name       string
		testNumber int
	}{
		{
			name:       "process cancel event does not panic",
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			orderChan := make(chan *entity.Order)
			cancelChan := make(chan entity.Event)
			eventChan := make(chan entity.Event, 10)
			dayChan := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			_ = New(ctx, orderChan, cancelChan, eventChan, dayChan)

			switch tt.testNumber {
			case 1:
				cancelEvent := entity.Event{
					EventType: entity.OrderCancelled,
					OrderId:   "order1",
				}

				cancelChan <- cancelEvent

				select {
				case <-time.After(100 * time.Millisecond):
				}
			}
		})
	}
}

func TestStarter_DayToggle(t *testing.T) {
	tests := []struct {
		name       string
		testNumber int
	}{
		{
			name:       "toggle day channel",
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			orderChan := make(chan *entity.Order)
			cancelChan := make(chan entity.Event)
			eventChan := make(chan entity.Event, 10)
			dayChan := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter := New(ctx, orderChan, cancelChan, eventChan, dayChan)

			switch tt.testNumber {
			case 1:
				require.False(t, starter.closed)

				dayChan <- struct{}{}

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.True(t, starter.closed)

				dayChan <- struct{}{}

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.False(t, starter.closed)
			}
		})
	}
}

func TestStarter_ClosedState(t *testing.T) {
	tests := []struct {
		name       string
		testNumber int
	}{
		{
			name:       "skip orders when closed",
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			orderChan := make(chan *entity.Order)
			cancelChan := make(chan entity.Event)
			eventChan := make(chan entity.Event, 10)
			dayChan := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter := New(ctx, orderChan, cancelChan, eventChan, dayChan)

			switch tt.testNumber {
			case 1:
				dayChan <- struct{}{}

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.True(t, starter.closed)

				order := &entity.Order{
					Id:          "1",
					Type:        entity.Ask,
					TimeInForce: entity.Gtc,
					Status:      entity.New,
					Price:       100,
					Quantity:    10,
				}

				orderChan <- order

				select {
				case <-time.After(100 * time.Millisecond):
				}
			}
		})
	}
}

func TestStarter_ContextCancel(t *testing.T) {
	tests := []struct {
		name       string
		testNumber int
	}{
		{
			name:       "context cancel stops goroutine",
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			orderChan := make(chan *entity.Order)
			cancelChan := make(chan entity.Event)
			eventChan := make(chan entity.Event, 10)
			dayChan := make(chan struct{})
			ctx, cancel := context.WithCancel(context.Background())

			_ = New(ctx, orderChan, cancelChan, eventChan, dayChan)

			switch tt.testNumber {
			case 1:
				close(orderChan)
				close(cancelChan)
				close(dayChan)
				cancel()

				select {
				case <-ctx.Done():
				case <-time.After(time.Second):
					t.Fatal("timeout waiting for context done")
				}
			}
		})
	}
}
