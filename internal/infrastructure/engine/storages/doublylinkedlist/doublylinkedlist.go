package doublylinkedlist

import (
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type DoublyLinkedList struct {
	Head   *BNode
	Tail   *BNode
	Size   int64
	Parent enginetypes.Container
}

type BNode struct {
	Prev   *BNode
	Next   *BNode
	Value  *enginetypes.EngineOrder
	Parent *DoublyLinkedList
}

func (this *DoublyLinkedList) SetParent(container enginetypes.Container) {
	this.Parent = container
}

func (this *DoublyLinkedList) Add(order *enginetypes.EngineOrder) {
	if order == nil {
		return
	}
	newNode := &BNode{
		Value:  order,
		Parent: this,
	}

	if this.Head == nil {
		this.Head = newNode
		this.Tail = newNode
	} else {
		this.Tail.Next = newNode
		newNode.Prev = this.Tail
		this.Tail = newNode
	}

	order.Parent = newNode

	this.Size++
}

func (this *DoublyLinkedList) Delete() {
	this.Size--
	if this.Size == 0 {
		this.Parent.Delete()
	}
}

func (this *DoublyLinkedList) GetNodes() func() *enginetypes.EngineOrder {
	curr := this.Head
	return func() *enginetypes.EngineOrder {
		if curr == nil {
			return nil
		}

		order := curr.Value

		curr = curr.Next

		return order
	}
}

func (this *BNode) Delete() {
	if this.Parent.Head == this {
		this.Parent.Head = this.Next
	}

	if this.Parent.Tail == this {
		this.Parent.Tail = this.Prev
	}

	if this.Next != nil {
		this.Next.Prev = this.Prev
	}

	if this.Prev != nil {
		this.Prev.Next = this.Next
	}

	this.Parent.Delete()
}

func (this *BNode) GetNextStorageNode() enginetypes.StorageNode {
	if this == nil {
		return nil
	}

	return this.Next
}

func (this *BNode) GetStorageNodeValue() *enginetypes.EngineOrder {
	if this == nil {
		return nil
	}
	return this.Value
}

func (this *DoublyLinkedList) GetFirstStorageNode() enginetypes.StorageNode {
	return this.Head
}
