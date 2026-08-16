package engine

import "github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"

type DoublyLinkedList struct {
	Head   *BNode
	Tail   *BNode
	Size   int64
	Parent *SkipNode
}

type BNode struct {
	prev   *BNode
	next   *BNode
	Value  *EngineOrder
	parent *DoublyLinkedList
}

type EngineOrder struct {
	Content *entity.Order
	parent  *BNode
}

func NewEngineOrder(order *entity.Order) *EngineOrder {
	return &EngineOrder{
		Content: order,
	}
}

func (this *EngineOrder) Delete() {
	this.parent.Delete()
}

func (this *DoublyLinkedList) Add(order *EngineOrder) {
	if order == nil {
		return
	}

	newNode := &BNode{
		Value:  order,
		parent: this,
	}
	order.parent = newNode

	if this.Head == nil {
		this.Head = newNode
		this.Tail = newNode
	} else {
		this.Tail.next = newNode
		newNode.prev = this.Tail
		this.Tail = newNode
	}

	this.Size++
}

func (this *DoublyLinkedList) Delete() {
	this.Size--
	if this.Size == 0 {
		this.Parent.Delete()
	}
}

func (this *BNode) Delete() {
	if this.parent.Head == this {
		this.parent.Head = this.next
	}

	if this.parent.Tail == this {
		this.parent.Tail = this.prev
	}

	if this.next != nil {
		this.next.prev = this.prev
	}

	if this.prev != nil {
		this.prev.next = this.next
	}

	this.parent.Delete()
}

type OrdersIterator struct {
	curr *BNode
}

func (this *DoublyLinkedList) NewIterator() *OrdersIterator {
	return &OrdersIterator{
		curr: this.Head,
	}
}

func (this *OrdersIterator) Next() *EngineOrder {
	order := this.curr.Value
	this.curr = this.curr.next
	return order
}
