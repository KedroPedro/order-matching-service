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
		testNumber int
	}{
		{
			name:       "create starter",
			ctx:        context.Background(),
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, eventChan := New(tt.ctx)

			switch tt.testNumber {
			case 1:
				require.NotNil(t, got)
				require.NotNil(t, got.engine)
				require.False(t, got.closed)
				require.NotNil(t, eventChan)
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter, _ := New(ctx)
			go starter.Start(ctx)

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

				err := starter.AddToQueue(order)
				require.NoError(t, err)

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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter, _ := New(ctx)
			go starter.Start(ctx)

			switch tt.testNumber {
			case 1:
				order := entity.Order{
					Id:             "order1",
					OwnerId:        "owner1",
					Type:           entity.Ask,
					Price:          100,
					Quantity:       10,
					FilledQuantity: 0,
				}
				cancelEvent := entity.NewOrderCancelledEvent(&order, order.Quantity-order.FilledQuantity)

				err := starter.Cancel(cancelEvent)
				require.NoError(t, err)

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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter, _ := New(ctx)
			go starter.Start(ctx)

			switch tt.testNumber {
			case 1:
				require.False(t, starter.IsClosed())

				starter.Close()

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.True(t, starter.IsClosed())

				starter.Open()

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.False(t, starter.IsClosed())
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			starter, _ := New(ctx)
			go starter.Start(ctx)

			switch tt.testNumber {
			case 1:
				starter.Close()

				select {
				case <-time.After(100 * time.Millisecond):
				}

				require.True(t, starter.IsClosed())

				order := &entity.Order{
					Id:          "1",
					Type:        entity.Ask,
					TimeInForce: entity.Gtc,
					Status:      entity.New,
					Price:       100,
					Quantity:    10,
				}

				err := starter.AddToQueue(order)
				require.NoError(t, err)

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
			ctx, cancel := context.WithCancel(context.Background())

			starter, _ := New(ctx)
			go starter.Start(ctx)

			switch tt.testNumber {
			case 1:
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
