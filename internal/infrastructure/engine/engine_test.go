package engine

import (
	"testing"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type OrderBookMock struct {
	mock.Mock
}

func (m *OrderBookMock) BestAskPrice() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *OrderBookMock) BestBidPrice() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *OrderBookMock) Match(order *enginetypes.EngineOrder) {
	m.Called(order)
}

func (m *OrderBookMock) Remove(orderId string) {
	m.Called(orderId)
}

func (m *OrderBookMock) Cancel(orderId string) {
	m.Called(orderId)
}

func (m *OrderBookMock) CancelDayOrders() {
	m.Called()
}

type StopOrderBookMock struct {
	mock.Mock
}

func (m *StopOrderBookMock) Add(order *enginetypes.EngineOrder) {
	m.Called(order)
}

func (m *StopOrderBookMock) GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder {
	args := m.Called(bestAskLevel, bestBidLevel)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*enginetypes.EngineOrder)
}

func (m *StopOrderBookMock) Cancel(orderId string) {
	m.Called(orderId)
}

func (m *StopOrderBookMock) CancelDayOrders() {
	m.Called()
}

func TestEngine_cancelOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
		orderId    string
	}{
		{
			name:    "cancel order in both books",
			orderId: "order1",
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Cancel", "order1").Return()
				stops.On("Cancel", "order1").Return()
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orders := &OrderBookMock{}
			stops := &StopOrderBookMock{}
			orderChan := make(<-chan *enginetypes.EngineOrder)
			cancelChan := make(<-chan string)
			endChan := make(chan<- struct{})
			dayChan := make(<-chan struct{})

			tt.setupMocks(orders, stops)

			this := New(t.Context(), orders, stops, orderChan, cancelChan, endChan, dayChan)
			this.cancelOrder(tt.orderId)

			orders.AssertCalled(t, "Cancel", tt.orderId)
			stops.AssertCalled(t, "Cancel", tt.orderId)
			orders.AssertExpectations(t)
			stops.AssertExpectations(t)
		})
	}
}

func TestEngine_close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
	}{
		{
			name: "close cancels day orders in both books",
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("CancelDayOrders").Return()
				stops.On("CancelDayOrders").Return()
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orders := &OrderBookMock{}
			stops := &StopOrderBookMock{}
			orderChan := make(<-chan *enginetypes.EngineOrder)
			cancelChan := make(<-chan string)
			endChan := make(chan<- struct{})
			dayChan := make(<-chan struct{})

			tt.setupMocks(orders, stops)

			this := New(t.Context(), orders, stops, orderChan, cancelChan, endChan, dayChan)
			this.close()

			orders.AssertCalled(t, "CancelDayOrders")
			stops.AssertCalled(t, "CancelDayOrders")
			orders.AssertExpectations(t)
			stops.AssertExpectations(t)
		})
	}
}

func TestEngine_processStopOrders(t *testing.T) {
	t.Parallel()

	stopOrder1 := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop1", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- *entity.Event),
	)

	tests := []struct {
		name           string
		setupMocks     func(orders *OrderBookMock, stops *StopOrderBookMock)
		assertBehavior func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock)
	}{
		{
			name: "no stop orders triggered",
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("BestAskPrice").Return(int64(100))
				orders.On("BestBidPrice").Return(int64(90))
				stops.On("GetStopOrders", int64(100), int64(90)).Return([]*enginetypes.EngineOrder{})
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.AssertNotCalled(t, "Match", mock.Anything)
			},
		},
		{
			name: "stop orders triggered and matched",
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("BestAskPrice").Return(int64(100))
				orders.On("BestBidPrice").Return(int64(90))
				stops.On("GetStopOrders", int64(100), int64(90)).Return([]*enginetypes.EngineOrder{stopOrder1})
				orders.On("Match", stopOrder1).Return()
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.AssertCalled(t, "Match", stopOrder1)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orders := &OrderBookMock{}
			stops := &StopOrderBookMock{}
			orderChan := make(<-chan *enginetypes.EngineOrder)
			cancelChan := make(<-chan string)
			endChan := make(chan<- struct{})
			dayChan := make(<-chan struct{})

			tt.setupMocks(orders, stops)

			this := New(t.Context(), orders, stops, orderChan, cancelChan, endChan, dayChan)
			this.processStopOrders()

			tt.assertBehavior(t, orders, stops)
			orders.AssertExpectations(t)
			stops.AssertExpectations(t)
		})
	}
}

func TestEngine_processOrder(t *testing.T) {
	t.Parallel()

	stopOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop1", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New, Stop: true},
		make(chan<- *entity.Event),
	)

	regularOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "reg1", Type: entity.Ask, TimeInForce: entity.Gtc, Status: entity.New},
		make(chan<- *entity.Event),
	)

	dayOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "day1", Type: entity.Ask, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- *entity.Event),
	)

	tests := []struct {
		name           string
		order          *enginetypes.EngineOrder
		closed         bool
		setupMocks     func(orders *OrderBookMock, stops *StopOrderBookMock)
		assertBehavior func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock, order *enginetypes.EngineOrder)
	}{
		{
			name:   "stop order - added to stop book",
			order:  stopOrder,
			closed: false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				stops.On("Add", stopOrder).Return()
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock, order *enginetypes.EngineOrder) {
				require.Equal(t, enginetypes.Pending, order.GetStatus())
				stops.AssertCalled(t, "Add", order)
				orders.AssertNotCalled(t, "Match")
			},
		},
		{
			name:   "regular order - matched",
			order:  regularOrder,
			closed: false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Match", regularOrder).Return()
				orders.On("BestAskPrice").Return(int64(0))
				orders.On("BestBidPrice").Return(int64(0))
				stops.On("GetStopOrders", int64(0), int64(0)).Return([]*enginetypes.EngineOrder{})
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock, order *enginetypes.EngineOrder) {
				require.Equal(t, enginetypes.New, order.GetStatus())
				orders.AssertCalled(t, "Match", order)
			},
		},
		{
			name:   "day order when closed - expired",
			order:  dayOrder,
			closed: true,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock, order *enginetypes.EngineOrder) {
				require.Equal(t, enginetypes.Expired, order.GetStatus())
				orders.AssertNotCalled(t, "Match")
			},
		},
		{
			name:   "day order when open - matched",
			order:  dayOrder,
			closed: false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Match", dayOrder).Return()
				orders.On("BestAskPrice").Return(int64(0))
				orders.On("BestBidPrice").Return(int64(0))
				stops.On("GetStopOrders", int64(0), int64(0)).Return([]*enginetypes.EngineOrder{})
			},
			assertBehavior: func(t *testing.T, orders *OrderBookMock, stops *StopOrderBookMock, order *enginetypes.EngineOrder) {
				require.Equal(t, enginetypes.New, order.GetStatus())
				orders.AssertCalled(t, "Match", order)
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			orders := &OrderBookMock{}
			stops := &StopOrderBookMock{}
			orderChan := make(<-chan *enginetypes.EngineOrder)
			cancelChan := make(<-chan string)
			endChan := make(chan<- struct{})
			dayChan := make(<-chan struct{})

			tt.setupMocks(orders, stops)

			this := New(t.Context(), orders, stops, orderChan, cancelChan, endChan, dayChan)
			if tt.closed {
				this.closed = true
			}
			this.processOrder(tt.order)

			tt.assertBehavior(t, orders, stops, tt.order)
			orders.AssertExpectations(t)
			stops.AssertExpectations(t)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	ordersMock := &OrderBookMock{}
	stopsMock := &StopOrderBookMock{}
	orderChan := make(<-chan *enginetypes.EngineOrder)
	cancelChan := make(<-chan string)
	endChan := make(chan<- struct{})
	dayChan := make(<-chan struct{})

	tests := []struct {
		name string
		want *Engine
	}{
		{
			name: "create engine with mocks",
			want: &Engine{
				orders:     ordersMock,
				stopOrders: stopsMock,
				closed:     false,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(t.Context(), ordersMock, stopsMock, orderChan, cancelChan, endChan, dayChan)

			require.NotNil(t, got)
			require.Equal(t, tt.want.orders, got.orders)
			require.Equal(t, tt.want.stopOrders, got.stopOrders)
			require.Equal(t, tt.want.closed, got.closed)
		})
	}
}
