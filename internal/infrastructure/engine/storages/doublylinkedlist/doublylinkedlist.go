package doublylinkedlist

import (
	engineinterfaces "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_interfaces"
	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
)

type DoublyLinkedList struct {
	Head   *BNode
	Tail   *BNode
	Size   int64
	Level  int64
	Parent engineinterfaces.Container
}

type BNode struct {
	Prev   *BNode
	Next   *BNode
	Value  *enginetypes.EngineOrder
	Parent *DoublyLinkedList
}

func (this *DoublyLinkedList) Add(order *enginetypes.EngineOrder) {
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
	if this.Size <= 0 {
		this.Parent.Delete()
	}
}

func (this *DoublyLinkedList) GetFirst() *BNode {
	return this.Head
}

func (this *BNode) Delete() {
	this.Parent.Delete()

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
}

func (this DoublyLinkedList) GetLevel() int64 {
	return this.Level
}

func (this *BNode) GetNext() *BNode {
	return this.Next
}
