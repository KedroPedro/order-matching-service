package doublylinkedlist

import "github.com/KedroPedro/order-matching-engine/internal/domain/entity"

type DoublyLinkedList struct {
	Head   *BNode
	Tail   *BNode
	Size   int64
	Parent any
}

type BNode struct {
	Prev    *BNode
	Next    *BNode
	Content *entity.Order
	Parent  *DoublyLinkedList
}

func (this *DoublyLinkedList) Add(order *entity.Order) {
	newNode := &BNode{
		Content: order,
		Parent:  this,
	}

	if this.Head == nil {
		this.Head = newNode
		this.Tail = newNode
	} else {
		this.Tail.Next = newNode
		newNode.Prev = this.Tail
		this.Tail = newNode
	}

	this.Size++
}

func (this *DoublyLinkedList) GetFirst() *BNode {
	return this.Head
}

func (this *BNode) Delete() {
	this.Parent.Size--

	if this.Next != nil {
		this.Next.Prev = this.Prev
	}

	if this.Prev != nil {
		this.Prev.Next = this.Next
	}
}
