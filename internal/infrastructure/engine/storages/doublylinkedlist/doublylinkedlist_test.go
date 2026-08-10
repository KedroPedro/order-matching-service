package doublylinkedlist_test

import (
	"testing"

	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/doublylinkedlist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDoublyLinkedList_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		orders     []*enginetypes.EngineOrder
		testNumber int
	}{
		{name: "zero orders", orders: []*enginetypes.EngineOrder{}, testNumber: 1},
		{name: "one order", orders: []*enginetypes.EngineOrder{{}}, testNumber: 2},
		{name: "two orders", orders: []*enginetypes.EngineOrder{{}, {}}, testNumber: 3},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			list := doublylinkedlist.DoublyLinkedList{}
			for _, order := range tt.orders {
				list.Add(order)
			}

			require.Equal(t, int64(len(tt.orders)), list.Size)

			switch tt.testNumber {
			case 1:
				require.Nil(t, list.Head)
				require.Nil(t, list.Tail)
			case 2:
				require.NotNil(t, list.Head)
				require.NotNil(t, list.Tail)
				assert.Same(t, tt.orders[0], list.Head.Value)
				assert.Same(t, tt.orders[0], list.Tail.Value)
				assert.Same(t, list.Head, list.Tail)
			case 3:
				require.NotNil(t, list.Head)
				require.NotNil(t, list.Tail)
				assert.Same(t, tt.orders[0], list.Head.Value)
				assert.Same(t, tt.orders[1], list.Tail.Value)
				assert.Same(t, list.Head.Next, list.Tail)
				assert.Same(t, list.Tail.Prev, list.Head)
				assert.NotSame(t, list.Head, list.Tail)
			}
		})
	}
}

func TestDoublyLinkedList_GetNodes(t *testing.T) {
	t.Parallel()

	orders := []*enginetypes.EngineOrder{{}, {}, {}}

	list := doublylinkedlist.DoublyLinkedList{}
	for i := range len(orders) {
		list.Add(orders[i])
	}

	tests := []struct {
		name string // description of this test case
		want func() *enginetypes.EngineOrder
	}{
		{name: "first iteration", want: list.GetNodes()},
		{name: "second iteration", want: list.GetNodes()},
		{name: "third iteration", want: list.GetNodes()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listOrders := make([]*enginetypes.EngineOrder, 0, len(orders))

			for curr := tt.want(); curr != nil; curr = tt.want() {
				listOrders = append(listOrders, curr)
			}

			require.Equal(t, len(orders), len(listOrders))

			for i, order := range orders {
				assert.Same(t, listOrders[i], order)
			}

		})
	}

}

type ListParent struct {
	mock.Mock
}

func (this *ListParent) Delete() {
	this.Called()
}

func TestDoublyLinkedList_Delete(t *testing.T) {
	t.Parallel()

	mockParent := ListParent{}

	mockParent.On("Delete").Return()

	list := doublylinkedlist.DoublyLinkedList{Size: 2, Parent: &mockParent}

	tests := []struct {
		name       string
		testNumber int
	}{
		{name: "delete element", testNumber: 1},
		{name: "delete last element", testNumber: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list.Delete()

			switch tt.testNumber {
			case 1:
				require.Equal(t, int64(1), list.Size)
				mockParent.AssertNotCalled(t, "Delete")
			case 2:
				require.Equal(t, int64(0), list.Size)
				mockParent.AssertCalled(t, "Delete")
			}
		})
	}
}

func TestBNode_Delete(t *testing.T) {
	t.Parallel()

	mockParent := ListParent{}

	mockParent.On("Delete").Return()

	list := doublylinkedlist.DoublyLinkedList{Size: 3, Parent: &mockParent}

	firstNode := doublylinkedlist.BNode{Parent: &list}
	secondNode := doublylinkedlist.BNode{Parent: &list, Prev: &firstNode}
	thirdNode := doublylinkedlist.BNode{Parent: &list, Prev: &secondNode}

	firstNode.Next = &secondNode
	secondNode.Next = &thirdNode

	list.Head = &firstNode
	list.Tail = &thirdNode

	tests := []struct {
		name       string // description of this test case
		testNumber int
	}{
		{name: "delete middle node", testNumber: 1},
		{name: "delete tail", testNumber: 2},
		{name: "delete head", testNumber: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.testNumber {
			case 1:
				secondNode.Delete()
				require.Same(t, list.Head, list.Tail.Prev)
				require.Same(t, list.Tail, list.Head.Next)
				assert.Equal(t, int64(2), list.Size)
				mockParent.AssertNotCalled(t, "Delete")
			case 2:
				thirdNode.Delete()
				require.Same(t, list.Head, list.Tail)
				assert.Equal(t, int64(1), list.Size)
				mockParent.AssertNotCalled(t, "Delete")
			case 3:
				firstNode.Delete()
				require.Nil(t, list.Head)
				require.Nil(t, list.Tail)
				assert.Equal(t, int64(0), list.Size)
				mockParent.AssertCalled(t, "Delete")
			}

		})
	}
}

func TestBNode_GetNext(t *testing.T) {
	t.Parallel()

	firstNode := &doublylinkedlist.BNode{}
	secondNode := &doublylinkedlist.BNode{Prev: firstNode}

	firstNode.Next = secondNode

	tests := []struct {
		name       string // description of this test case
		want       *doublylinkedlist.BNode
		testNumber int
	}{
		{name: "get second node", want: secondNode, testNumber: 1},
		{name: "get nil", want: nil, testNumber: 2},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch tt.testNumber {
			case 1:
				require.Same(t, firstNode.GetNextStorageNode(), tt.want)
			case 2:
				require.Same(t, secondNode.GetNextStorageNode(), tt.want)
			}
		})
	}
}
