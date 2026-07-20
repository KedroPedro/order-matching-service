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
		orders     *OrderBookMock
		stops      *StopOrderBookMock
		orderChan  <-chan *enginetypes.EngineOrder
		cancelChan <-chan string
		endChan    chan<- struct{}
		dayChan    <-chan struct{}
		orderId    string
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
		testNumber int
	}{
		{
			name:       "cancel order in both books",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			orderId:    "order1",
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Cancel", "order1").Return()
				stops.On("Cancel", "order1").Return()
			},
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setupMocks(tt.orders, tt.stops)

			this := New(t.Context(), tt.orders, tt.stops, tt.orderChan, tt.cancelChan, tt.endChan, tt.dayChan)
			this.cancelOrder(tt.orderId)

			tt.orders.AssertCalled(t, "Cancel", tt.orderId)
			tt.stops.AssertCalled(t, "Cancel", tt.orderId)
			tt.orders.AssertExpectations(t)
			tt.stops.AssertExpectations(t)
		})
	}
}

func TestEngine_close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		orders     *OrderBookMock
		stops      *StopOrderBookMock
		orderChan  <-chan *enginetypes.EngineOrder
		cancelChan <-chan string
		endChan    chan<- struct{}
		dayChan    <-chan struct{}
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
		testNumber int
	}{
		{
			name:       "close cancels day orders in both books",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("CancelDayOrders").Return()
				stops.On("CancelDayOrders").Return()
			},
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setupMocks(tt.orders, tt.stops)

			this := New(t.Context(), tt.orders, tt.stops, tt.orderChan, tt.cancelChan, tt.endChan, tt.dayChan)
			this.close()

			tt.orders.AssertCalled(t, "CancelDayOrders")
			tt.stops.AssertCalled(t, "CancelDayOrders")
			tt.orders.AssertExpectations(t)
			tt.stops.AssertExpectations(t)
		})
	}
}

func TestEngine_processStopOrders(t *testing.T) {
	t.Parallel()

	stopOrder1 := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop1", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- entity.Event),
	)

	tests := []struct {
		name       string
		orders     *OrderBookMock
		stops      *StopOrderBookMock
		orderChan  <-chan *enginetypes.EngineOrder
		cancelChan <-chan string
		endChan    chan<- struct{}
		dayChan    <-chan struct{}
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
		testNumber int
	}{
		{
			name:       "no stop orders triggered",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("BestAskPrice").Return(int64(100))
				orders.On("BestBidPrice").Return(int64(90))
				stops.On("GetStopOrders", int64(100), int64(90)).Return([]*enginetypes.EngineOrder{})
			},
			testNumber: 1,
		},
		{
			name:       "stop orders triggered and matched",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("BestAskPrice").Return(int64(100))
				orders.On("BestBidPrice").Return(int64(90))
				stops.On("GetStopOrders", int64(100), int64(90)).Return([]*enginetypes.EngineOrder{stopOrder1})
				orders.On("Match", stopOrder1).Return()
			},
			testNumber: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setupMocks(tt.orders, tt.stops)

			this := New(t.Context(), tt.orders, tt.stops, tt.orderChan, tt.cancelChan, tt.endChan, tt.dayChan)
			this.processStopOrders()

			switch tt.testNumber {
			case 1:
				tt.orders.AssertNotCalled(t, "Match", mock.Anything)
			case 2:
				tt.orders.AssertCalled(t, "Match", stopOrder1)
			}

			tt.orders.AssertExpectations(t)
			tt.stops.AssertExpectations(t)
		})
	}
}

func TestEngine_processOrder(t *testing.T) {
	t.Parallel()

	stopOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop1", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New, Stop: true},
		make(chan<- entity.Event),
	)

	regularOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "reg1", Type: entity.Ask, TimeInForce: entity.Gtc, Status: entity.New},
		make(chan<- entity.Event),
	)

	dayOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "day1", Type: entity.Ask, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- entity.Event),
	)

	tests := []struct {
		name       string
		orders     *OrderBookMock
		stops      *StopOrderBookMock
		orderChan  <-chan *enginetypes.EngineOrder
		cancelChan <-chan string
		endChan    chan<- struct{}
		dayChan    <-chan struct{}
		order      *enginetypes.EngineOrder
		closed     bool
		setupMocks func(orders *OrderBookMock, stops *StopOrderBookMock)
		testNumber int
	}{
		{
			name:       "stop order - added to stop book",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			order:      stopOrder,
			closed:     false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				stops.On("Add", stopOrder).Return()
			},
			testNumber: 1,
		},
		{
			name:       "regular order - matched",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			order:      regularOrder,
			closed:     false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Match", regularOrder).Return()
				orders.On("BestAskPrice").Return(int64(0))
				orders.On("BestBidPrice").Return(int64(0))
				stops.On("GetStopOrders", int64(0), int64(0)).Return([]*enginetypes.EngineOrder{})
			},
			testNumber: 2,
		},
		{
			name:       "day order when closed - expired",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			order:      dayOrder,
			closed:     true,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
			},
			testNumber: 3,
		},
		{
			name:       "day order when open - matched",
			orders:     &OrderBookMock{},
			stops:      &StopOrderBookMock{},
			orderChan:  make(<-chan *enginetypes.EngineOrder),
			cancelChan: make(<-chan string),
			endChan:    make(chan<- struct{}),
			dayChan:    make(<-chan struct{}),
			order:      dayOrder,
			closed:     false,
			setupMocks: func(orders *OrderBookMock, stops *StopOrderBookMock) {
				orders.On("Match", dayOrder).Return()
				orders.On("BestAskPrice").Return(int64(0))
				orders.On("BestBidPrice").Return(int64(0))
				stops.On("GetStopOrders", int64(0), int64(0)).Return([]*enginetypes.EngineOrder{})
			},
			testNumber: 4,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setupMocks(tt.orders, tt.stops)

			this := New(t.Context(), tt.orders, tt.stops, tt.orderChan, tt.cancelChan, tt.endChan, tt.dayChan)
			if tt.closed {
				this.closed = true
			}
			this.processOrder(tt.order)

			switch tt.testNumber {
			case 1:
				require.Equal(t, enginetypes.Pending, tt.order.GetStatus())
				tt.stops.AssertCalled(t, "Add", tt.order)
				tt.orders.AssertNotCalled(t, "Match")
			case 2, 4:
				require.Equal(t, enginetypes.New, tt.order.GetStatus())
				tt.orders.AssertCalled(t, "Match", tt.order)
			case 3:
				require.Equal(t, enginetypes.Expired, tt.order.GetStatus())
				tt.orders.AssertNotCalled(t, "Match")
			}

			tt.orders.AssertExpectations(t)
			tt.stops.AssertExpectations(t)
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
		name       string
		orders     *OrderBookMock
		stops      *StopOrderBookMock
		orderChan  <-chan *enginetypes.EngineOrder
		cancelChan <-chan string
		endChan    chan<- struct{}
		dayChan    <-chan struct{}
		want       *Engine
		testNumber int
	}{
		{
			name:       "create engine with mocks",
			orders:     ordersMock,
			stops:      stopsMock,
			orderChan:  orderChan,
			cancelChan: cancelChan,
			endChan:    endChan,
			dayChan:    dayChan,
			want: &Engine{
				orders:     ordersMock,
				stopOrders: stopsMock,
				closed:     false,
			},
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := New(t.Context(), tt.orders, tt.stops, tt.orderChan, tt.cancelChan, tt.endChan, tt.dayChan)

			switch tt.testNumber {
			case 1:
				require.NotNil(t, got)
				require.Equal(t, tt.want.orders, got.orders)
				require.Equal(t, tt.want.stopOrders, got.stopOrders)
				require.Equal(t, tt.want.closed, got.closed)
			}
		})
	}
}
