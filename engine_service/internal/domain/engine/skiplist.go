package engine

import (
	"math/rand/v2"
)

type ListType string

const (
	DescendingList ListType = "desc"
	AscendingList  ListType = "asc"
)

const (
	maxNodeHeight              = 16
	nodesMapDefaultSize        = 128
	priceLevelSliceDefaultSize = 16
)

type SkipList struct {
	Head             *SkipNode
	Size             int64
	TotalQuantity    int64
	compare          func(i, j int64) bool
	nodes            map[int64]*SkipNode
	maxHeight        int16
	priceLevelBuffer []*PriceLevel
}

type SkipNode struct {
	forward [maxNodeHeight]*SkipNode
	Parent  *SkipList
	Level   *PriceLevel
}

type PriceLevel struct {
	Price         int64
	TotalQuantity int64
	Orders        *DoublyLinkedList
	Parent        *SkipNode
}

func NewSkipList(listType ListType) *SkipList {
	newSkipList := &SkipList{
		Size:             0,
		TotalQuantity:    0,
		maxHeight:        1,
		nodes:            make(map[int64]*SkipNode, nodesMapDefaultSize),
		priceLevelBuffer: make([]*PriceLevel, 0, priceLevelSliceDefaultSize),
	}

	newSkipList.Head = &SkipNode{
		Parent: newSkipList,
		Level:  nil,
	}

	switch listType {
	case DescendingList:
		newSkipList.compare = func(i, j int64) bool { return i > j }
	case AscendingList:
		newSkipList.compare = func(i, j int64) bool { return i < j }
	}

	return newSkipList
}

func (this *SkipList) Add(order *EngineOrder) {
	if order == nil {
		return
	}

	if node, ok := this.nodes[order.Content.Price]; ok {
		node.Level.Orders.Add(order)
		node.Level.TotalQuantity += order.Content.GetUnfilledQuantity()
		return
	}

	var newNode *SkipNode
	newNode = &SkipNode{
		Parent: this,
		Level: &PriceLevel{
			Price:         order.Content.Price,
			TotalQuantity: order.Content.GetUnfilledQuantity(),
			Orders:        &DoublyLinkedList{},
			Parent:        newNode,
		},
	}
	newNode.Level.Orders.Add(order)

	nodeHeight := getRandomHeight()
	if this.maxHeight < nodeHeight {
		this.maxHeight = nodeHeight
	}

	curr := this.Head

	for i := this.maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && this.compare(curr.forward[i].Level.Price, order.Content.Price) {
			curr = curr.forward[i]
		}

		if i <= nodeHeight {
			newNode.forward[i] = curr.forward[i]
			curr.forward[i] = newNode
		}
	}

	this.nodes[order.Content.Price] = newNode
	this.Size++
}

func (this *SkipList) GetRange(quantity int64, price int64) []*PriceLevel {
	this.priceLevelBuffer = this.priceLevelBuffer[:0]

	if this.Head.forward[0] != nil {
		curr := this.Head.forward[0]
		for quantity > 0 && curr != nil {

			if !this.compare(curr.Level.Price, price) && curr.Level.Price != price {
				break
			}
			quantity -= curr.Level.TotalQuantity
			this.priceLevelBuffer = append(this.priceLevelBuffer, curr.Level)

			curr = curr.forward[0]
		}
	}

	return this.priceLevelBuffer
}

func (this *SkipList) GetFirst() *PriceLevel {
	if this.Head.forward[0] != nil {
		return this.Head.forward[0].Level
	}

	return nil
}

func (this *SkipList) Delete(level int64) {
	target, ok := this.nodes[level]
	if !ok {
		return
	}

	curr := this.Head

	for i := this.maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && this.compare(curr.forward[i].Level.Price, level) {
			curr = curr.forward[i]
		}

		if curr.forward[i] == target {
			curr.forward[i] = target.forward[i]
		}
	}

	this.Size--

	for this.maxHeight > 1 && this.Head.forward[this.maxHeight-1] == nil {
		this.maxHeight--
	}
}

func (this *SkipNode) Delete() {
	this.Parent.Delete(this.Level.Price)
}

func (this *PriceLevel) Delete() {
	if this.TotalQuantity <= 0 {
		this.Parent.Delete()
	}
}

func getRandomHeight() int16 {
	var height int16 = 1

	for height < maxNodeHeight && rand.Int32N(4) == 0 {
		height++
	}
	return height
}
