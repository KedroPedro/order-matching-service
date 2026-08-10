package engine_test

import (
	"context"
	"testing"
	"time"

	eventbatch "github.com/KedroPedro/order-matching-engine/internal/application/event_handler/event_batch"
	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOrderBook struct {
	mock.Mock
}

func (m *MockOrderBook) BestAskPrice() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockOrderBook) BestBidPrice() int64 {
	args := m.Called()
	return args.Get(0).(int64)
}

func (m *MockOrderBook) Match(order *enginetypes.EngineOrder, batch *eventbatch.EventBatch) {
	m.Called(order, batch)
}

func (m *MockOrderBook) Remove(orderId string) {
	m.Called(orderId)
}

func (m *MockOrderBook) Cancel(orderId string, batch *eventbatch.EventBatch) {
	m.Called(orderId, batch)
}

func (m *MockOrderBook) CancelDayOrders(batch *eventbatch.EventBatch) {
	m.Called(batch)
}

type MockStopOrderBook struct {
	mock.Mock
}

func (m *MockStopOrderBook) Add(order *enginetypes.EngineOrder) {
	m.Called(order)
}

func (m *MockStopOrderBook) GetStopOrders(bestAskLevel, bestBidLevel int64) []*enginetypes.EngineOrder {
	args := m.Called(bestAskLevel, bestBidLevel)
	return args.Get(0).([]*enginetypes.EngineOrder)
}

func (m *MockStopOrderBook) Cancel(orderId string, batch *eventbatch.EventBatch) {
	m.Called(orderId, batch)
}

func (m *MockStopOrderBook) CancelDayOrders(batch *eventbatch.EventBatch) {
	m.Called(batch)
}

func createTestOrder(id string, orderType enginetypes.EngineOrderType, tif entity.OrderTimeInForce, price, quantity int64) *enginetypes.EngineOrder {
	var entityOrderType entity.OrderType
	if orderType == enginetypes.Ask {
		entityOrderType = entity.Ask
	} else {
		entityOrderType = entity.Bid
	}

	order := &entity.Order{
		Id:             id,
		OwnerId:        "test-owner",
		Type:           entityOrderType,
		Price:          price,
		Quantity:       quantity,
		FilledQuantity: 0,
		Class:          entity.Limit,
		TimeInForce:    tif,
		Status:         entity.New,
		CreatedAt:      time.Now(),
		Reserve:        price * quantity,
		Stop:           false,
	}

	return enginetypes.NewEngineOrder(order, nil)
}

func createStopOrder(id string, orderType enginetypes.EngineOrderType, tif entity.OrderTimeInForce, price, quantity int64) *enginetypes.EngineOrder {
	var entityOrderType entity.OrderType
	if orderType == enginetypes.Ask {
		entityOrderType = entity.Ask
	} else {
		entityOrderType = entity.Bid
	}

	order := &entity.Order{
		Id:             id,
		OwnerId:        "test-owner",
		Type:           entityOrderType,
		Price:          price,
		Quantity:       quantity,
		FilledQuantity: 0,
		Class:          entity.Limit,
		TimeInForce:    tif,
		Status:         entity.New,
		CreatedAt:      time.Now(),
		Reserve:        price * quantity,
		Stop:           true,
	}

	return enginetypes.NewEngineOrder(order, nil)
}

func TestEngine_ProcessOrder_NormalOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockOrders.On("Match", mock.Anything, mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	order := createTestOrder("order-1", enginetypes.Bid, entity.Gtc, 100, 50)
	orderChan <- order

	select {
	case batch := <-eventBatchChan:
		require.Equal(t, 2, batch.Len())
		batch.Release()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event batch")
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}

func TestEngine_ProcessOrder_StopOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockStops.On("Add", mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	stopOrder := createStopOrder("stop-1", enginetypes.Bid, entity.Gtc, 105, 50)
	orderChan <- stopOrder

	select {
	case batch := <-eventBatchChan:
		require.Equal(t, 1, batch.Len())
		batch.Release()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event batch")
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}

func TestEngine_CancelOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockOrders.On("Cancel", "order-1", mock.Anything).Return()
	mockStops.On("Cancel", "order-1", mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	cancelChan <- "order-1"

	select {
	case batch := <-eventBatchChan:
		require.Equal(t, 0, batch.Len())
		batch.Release()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event batch")
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}

func TestEngine_Close(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockOrders.On("CancelDayOrders", mock.Anything).Return()
	mockStops.On("CancelDayOrders", mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	dayChan <- struct{}{}

	select {
	case batch := <-eventBatchChan:
		require.Equal(t, 0, batch.Len())
		batch.Release()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event batch")
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}

func TestEngine_ProcessStopOrders(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	stopOrder := createStopOrder("stop-1", enginetypes.Bid, entity.Gtc, 105, 50)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockOrders.On("Match", mock.Anything, mock.Anything).Return()
	mockStops.On("Add", mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{stopOrder})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	normalOrder := createTestOrder("order-1", enginetypes.Bid, entity.Gtc, 100, 50)
	orderChan <- normalOrder

	var batches []*eventbatch.EventBatch
	timeout := time.After(200 * time.Millisecond)
	for len(batches) < 2 {
		select {
		case batch := <-eventBatchChan:
			batches = append(batches, batch)
		case <-timeout:
			t.Fatalf("timeout waiting for event batches, got %d", len(batches))
		}
	}

	require.Equal(t, 2, len(batches))
	for _, batch := range batches {
		batch.Release()
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}

func TestEngine_ProcessOrder_DayOrderWhenClosed(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockOrders := new(MockOrderBook)
	mockStops := new(MockStopOrderBook)

	eventBatchChan := make(chan *eventbatch.EventBatch, 10)
	orderChan := make(chan *enginetypes.EngineOrder, 10)
	cancelChan := make(chan string, 10)
	endChan := make(chan struct{}, 1)
	dayChan := make(chan struct{}, 1)

	mockOrders.On("BestAskPrice").Return(int64(100))
	mockOrders.On("BestBidPrice").Return(int64(99))
	mockOrders.On("CancelDayOrders", mock.Anything).Return()
	mockStops.On("CancelDayOrders", mock.Anything).Return()
	mockStops.On("GetStopOrders", int64(100), int64(99)).Return([]*enginetypes.EngineOrder{})

	engine.New(ctx, mockOrders, mockStops, eventBatchChan, orderChan, cancelChan, endChan, dayChan)

	dayChan <- struct{}{}

	select {
	case batch := <-eventBatchChan:
		batch.Release()
	case <-time.After(100 * time.Millisecond):
	}

	dayOrder := createTestOrder("day-1", enginetypes.Bid, entity.Day, 100, 50)
	orderChan <- dayOrder

	select {
	case batch := <-eventBatchChan:
		require.Equal(t, 2, batch.Len())
		batch.Release()
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for event batch")
	}

	close(orderChan)
	close(cancelChan)
	close(dayChan)

	select {
	case <-endChan:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for engine shutdown")
	}

	mockOrders.AssertExpectations(t)
	mockStops.AssertExpectations(t)
}
