package orderbook_test

import (
	"testing"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/orderbook"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type ParentMock struct {
	mock.Mock
}

func (this *ParentMock) Delete() {
	this.Called()
}

type BookStorageMock struct {
	mock.Mock
	Added []*enginetypes.EngineOrder
}

func (this *BookStorageMock) Add(order *enginetypes.EngineOrder) {
	this.Added = append(this.Added, order)
	this.Called()
}

func (this *BookStorageMock) GetRange(quantity int64) []*enginetypes.PriceLevel {
	return nil
}

func (this *BookStorageMock) GetFirst() *enginetypes.PriceLevel {
	return nil
}

func (this *BookStorageMock) Delete(level int64) {}

func TestOrderBook_Add(t *testing.T) {
	t.Parallel()

	firstOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok},
		make(chan<- *entity.Event),
	)

	secondOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "2", Type: entity.Bid, TimeInForce: entity.Day},
		make(chan<- *entity.Event),
	)

	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		orders     []*enginetypes.EngineOrder
		testNumber int
	}{
		{name: "add nil order", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{nil}, testNumber: 1},
		{name: "add bid order", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{secondOrder}, testNumber: 2},
		{name: "add ask order", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{firstOrder}, testNumber: 3},
		{name: "add two orders", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{firstOrder, secondOrder}, testNumber: 4},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.askStorage.On("Add").Return()
			tt.bidStorage.On("Add").Return()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			for _, order := range tt.orders {
				book.Add(order)
			}

			switch tt.testNumber {
			case 1:
				tt.askStorage.AssertNotCalled(t, "Add")
				tt.bidStorage.AssertNotCalled(t, "Add")

			case 2:
				tt.askStorage.AssertNotCalled(t, "Add")
				tt.bidStorage.AssertCalled(t, "Add")

			case 3:
				tt.askStorage.AssertCalled(t, "Add")
				tt.bidStorage.AssertNotCalled(t, "Add")

			case 4:
				tt.askStorage.AssertCalled(t, "Add")
				tt.bidStorage.AssertCalled(t, "Add")
			}
		})
	}
}

func TestOrderBook_Remove(t *testing.T) {
	t.Parallel()

	firstOrder := enginetypes.NewEngineOrder(
		&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok},
		make(chan<- *entity.Event),
	)

	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		orders     []*enginetypes.EngineOrder
		orderId    string
		testNumber int
	}{
		{name: "remove non-existent order in empty book", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{nil}, orderId: "1", testNumber: 1},
		{name: "remove existent order", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{firstOrder}, orderId: "1", testNumber: 2},
		{name: "remove non-existent order in filled book", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{firstOrder}, orderId: "2", testNumber: 3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {

			parent := &ParentMock{}

			parent.On("Delete").Return()

			tt.askStorage.On("Add").Return()
			tt.bidStorage.On("Add").Return()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			for _, order := range tt.orders {
				if order != nil {
					order.Parent = parent
				}
				book.Add(order)
			}

			book.Remove(tt.orderId)

			switch tt.testNumber {
			case 1:
				parent.AssertNotCalled(t, "Delete")

			case 2:
				parent.AssertCalled(t, "Delete")

			case 3:
				parent.AssertNotCalled(t, "Delete")

			}

		})
	}
}

func TestOrderBook_CancelDayOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		orders     []*enginetypes.EngineOrder
		testNumber int
	}{
		{name: "no day orders", testNumber: 1, askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok, Status: entity.New},
				make(chan<- *entity.Event),
			)},
		},
		{name: "with day order", testNumber: 2, askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok, Status: entity.New},
				make(chan<- *entity.Event),
			),
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "2", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New},
				make(chan<- *entity.Event),
			)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parent := &ParentMock{}

			parent.On("Delete").Return()

			tt.askStorage.On("Add").Return()
			tt.bidStorage.On("Add").Return()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			for _, order := range tt.orders {
				order.Parent = parent
				book.Add(order)
			}

			book.CancelDayOrders()

			switch tt.testNumber {
			case 1:
				require.Equal(t, enginetypes.New, tt.orders[0].GetStatus())

			case 2:
				require.Equal(t, enginetypes.New, tt.orders[0].GetStatus())
				require.Equal(t, enginetypes.Expired, tt.orders[1].GetStatus())

			}
		})
	}
}

func TestOrderBook_Cancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		orders     []*enginetypes.EngineOrder
		orderId    string
		testNumber int
	}{
		{name: "cancel non-existent order", testNumber: 1, orderId: "-1", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok, Status: entity.New},
				make(chan<- *entity.Event),
			)},
		},
		{name: "cancel existent order", testNumber: 2, orderId: "1", askStorage: &BookStorageMock{}, bidStorage: &BookStorageMock{}, orders: []*enginetypes.EngineOrder{
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "1", Type: entity.Ask, TimeInForce: entity.Fok, Status: entity.New},
				make(chan<- *entity.Event),
			),
			enginetypes.NewEngineOrder(
				&entity.Order{Id: "2", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New},
				make(chan<- *entity.Event),
			)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parent := &ParentMock{}

			parent.On("Delete").Return()

			tt.askStorage.On("Add").Return()
			tt.bidStorage.On("Add").Return()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			for _, order := range tt.orders {
				order.Parent = parent
				book.Add(order)
			}

			book.Cancel(tt.orderId)

			switch tt.testNumber {
			case 1:
				require.Equal(t, enginetypes.New, tt.orders[0].GetStatus())

			case 2:
				require.Equal(t, enginetypes.Canceled, tt.orders[0].GetStatus())
				require.Equal(t, enginetypes.New, tt.orders[1].GetStatus())

			}
		})
	}
}

func TestOrderBook_GetStopOrders(t *testing.T) {
	t.Parallel()

	stopOrder1 := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop1", Type: entity.Bid, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- *entity.Event),
	)

	stopOrder2 := enginetypes.NewEngineOrder(
		&entity.Order{Id: "stop2", Type: entity.Ask, TimeInForce: entity.Day, Status: entity.New},
		make(chan<- *entity.Event),
	)

	parent := &ParentMock{}
	parent.On("Delete").Return()

	stopOrder1.Parent = parent
	stopOrder2.Parent = parent

	tests := []struct {
		name         string
		askStorage   *BookStorageMock
		bidStorage   *BookStorageMock
		bestAskLevel int64
		bestBidLevel int64
		setupMocks   func(ask, bid *BookStorageMock)
		want         []*enginetypes.EngineOrder
		testNumber   int
	}{
		{
			name:         "empty book - no stop orders",
			askStorage:   &BookStorageMock{},
			bidStorage:   &BookStorageMock{},
			bestAskLevel: 150,
			bestBidLevel: 50,
			setupMocks: func(ask, bid *BookStorageMock) {
			},
			want:       []*enginetypes.EngineOrder{},
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.setupMocks(tt.askStorage, tt.bidStorage)

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			got := book.GetStopOrders(tt.bestAskLevel, tt.bestBidLevel)

			switch tt.testNumber {
			case 1:
				require.Empty(t, got)
			}
		})
	}
}

func TestOrderBook_BestBidPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		want       int64
		testNumber int
	}{
		{
			name:       "empty bid book",
			askStorage: &BookStorageMock{},
			bidStorage: &BookStorageMock{},
			want:       0,
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			got := book.BestBidPrice()

			require.Equal(t, tt.want, got)
		})
	}
}

func TestOrderBook_BestAskPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		askStorage *BookStorageMock
		bidStorage *BookStorageMock
		want       int64
		testNumber int
	}{
		{
			name:       "empty ask book",
			askStorage: &BookStorageMock{},
			bidStorage: &BookStorageMock{},
			want:       0,
			testNumber: 1,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			book := orderbook.New(tt.askStorage, tt.bidStorage)

			got := book.BestAskPrice()

			require.Equal(t, tt.want, got)
		})
	}
}
