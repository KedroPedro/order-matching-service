package skiplist_test

import (
	"testing"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/skiplist"
	"github.com/stretchr/testify/require"
)

func TestSkipList_Add(t *testing.T) {
	t.Parallel()

	firstOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 1}, nil)
	secondOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 2}, nil)
	thirdOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 3}, nil)

	tests := []struct {
		name       string
		orders     []*enginetypes.EngineOrder
		testNumber int
	}{
		{name: "add nil order", orders: nil, testNumber: 1},
		{name: "add one order ascending", orders: []*enginetypes.EngineOrder{firstOrder}, testNumber: 2},
		{name: "add two order ascending", orders: []*enginetypes.EngineOrder{firstOrder, secondOrder}, testNumber: 3},
		{name: "add three order ascending", orders: []*enginetypes.EngineOrder{firstOrder, secondOrder, thirdOrder}, testNumber: 4},
		{name: "add one order descending", orders: []*enginetypes.EngineOrder{firstOrder}, testNumber: 5},
		{name: "add two order descending", orders: []*enginetypes.EngineOrder{firstOrder, secondOrder}, testNumber: 6},
		{name: "add three order descending", orders: []*enginetypes.EngineOrder{firstOrder, secondOrder, thirdOrder}, testNumber: 7},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var currType skiplist.ListType
			if tt.testNumber <= 4 {
				currType = skiplist.AscendingList
			} else {
				currType = skiplist.DescendingList
			}

			list := skiplist.New(currType)

			for _, order := range tt.orders {
				list.Add(order)
			}

			levels := list.GetRange(1, 4)

			switch tt.testNumber {
			case 1:
				require.Empty(t, levels)
			case 2:
				require.Equal(t, len(levels), 1)
				require.Same(t, levels[0].GetFirst(), firstOrder)

			case 3:
				require.Same(t, levels[0].GetFirst(), firstOrder)
				require.Same(t, levels[1].GetFirst(), secondOrder)
				require.Equal(t, len(levels), 2)

			case 4:
				require.Same(t, levels[0].GetFirst(), firstOrder)
				require.Same(t, levels[1].GetFirst(), secondOrder)
				require.Same(t, levels[2].GetFirst(), thirdOrder)
				require.Equal(t, len(levels), 3)

			case 5:
				require.Same(t, levels[0].GetFirst(), firstOrder)
				require.Equal(t, len(levels), 1)

			case 6:
				require.Same(t, levels[0].GetFirst(), secondOrder)
				require.Same(t, levels[1].GetFirst(), firstOrder)
				require.Equal(t, len(levels), 2)

			case 7:
				require.Same(t, levels[0].GetFirst(), thirdOrder)
				require.Same(t, levels[1].GetFirst(), secondOrder)
				require.Same(t, levels[2].GetFirst(), firstOrder)
				require.Equal(t, len(levels), 3)

			}

		})
	}
}

func TestSkipList_GetRange(t *testing.T) {
	t.Parallel()

	firstOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 1, Quantity: 1}, nil)
	secondOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 2, Quantity: 2}, nil)
	thirdOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 3, Quantity: 3}, nil)

	tests := []struct {
		name       string
		listType   skiplist.ListType
		orders     []*enginetypes.EngineOrder
		quantity   int64
		testNumber int
	}{
		{name: "empty list, quantity 0", listType: skiplist.AscendingList, orders: nil, quantity: 0, testNumber: 1},
		{name: "empty list, quantity 5", listType: skiplist.AscendingList, orders: nil, quantity: 5, testNumber: 2},
		{name: "ascending, quantity 1", listType: skiplist.AscendingList, orders: []*enginetypes.EngineOrder{firstOrder}, quantity: 1, testNumber: 3},
		{name: "ascending, quantity 2", listType: skiplist.AscendingList, orders: []*enginetypes.EngineOrder{firstOrder, secondOrder}, quantity: 2, testNumber: 4},
		{name: "ascending, quantity exceeds size", listType: skiplist.AscendingList, orders: []*enginetypes.EngineOrder{firstOrder, secondOrder, thirdOrder}, quantity: 5, testNumber: 5},
		{name: "descending, quantity 1", listType: skiplist.DescendingList, orders: []*enginetypes.EngineOrder{firstOrder, secondOrder}, quantity: 1, testNumber: 6},
		{name: "descending, quantity 2", listType: skiplist.DescendingList, orders: []*enginetypes.EngineOrder{firstOrder, secondOrder, thirdOrder}, quantity: 2, testNumber: 7},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			list := skiplist.New(tt.listType)

			for _, order := range tt.orders {
				list.Add(order)
			}

			got := list.GetRange(tt.quantity, 4)

			switch tt.testNumber {
			case 1, 2:
				require.Empty(t, got)

			case 3:
				require.Len(t, got, 1)
				require.Same(t, got[0].GetFirst(), firstOrder)

			case 4:
				require.Len(t, got, 2)
				require.Same(t, got[0].GetFirst(), firstOrder)
				require.Same(t, got[1].GetFirst(), secondOrder)

			case 5:
				require.Len(t, got, 3)
				require.Same(t, got[0].GetFirst(), firstOrder)
				require.Same(t, got[1].GetFirst(), secondOrder)
				require.Same(t, got[2].GetFirst(), thirdOrder)

			case 6:
				require.Len(t, got, 1)
				require.Same(t, got[0].GetFirst(), secondOrder)

			case 7:
				require.Same(t, got[0].GetFirst(), thirdOrder)
				require.Len(t, got, 1)
			}
		})
	}
}

func TestSkipList_Delete(t *testing.T) {
	firstOrder := enginetypes.NewEngineOrder(&entity.Order{Price: 1}, nil)

	tests := []struct {
		name       string
		orders     []*enginetypes.EngineOrder
		level      int64
		testNumber int
	}{
		{name: "delete from empty list", level: 1, testNumber: 1, orders: []*enginetypes.EngineOrder{}},
		{name: "delete existent level", level: 1, testNumber: 2, orders: []*enginetypes.EngineOrder{firstOrder}},
		{name: "delete non-existent level", level: 123, testNumber: 3, orders: []*enginetypes.EngineOrder{firstOrder}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			list := skiplist.New(skiplist.AscendingList)

			for _, order := range tt.orders {
				list.Add(order)
			}

			list.Delete(tt.level)

			switch tt.testNumber {
			case 1:
				require.Equal(t, int64(0), list.Size)
				require.Nil(t, list.GetFirst())

			case 2:
				require.Equal(t, int64(0), list.Size)
				require.Nil(t, list.GetFirst())

			case 3:
				require.Equal(t, int64(1), list.Size)
				require.Same(t, firstOrder, list.GetFirst().GetFirst())
			}
		})
	}
}
