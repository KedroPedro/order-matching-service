package skiplist

import (
	"math/rand/v2"

	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/doublylinkedlist"
)

type ListType string

const (
	DescendingList ListType = "desc"
	AscendingList  ListType = "asc"
)

const (
	maxNodeHeight = 16
)

type SkipList struct {
	Head          *SkipNode
	Size          int64
	TotalQuantity int64
	compare       func(i, j int64) bool
	nodes         map[int64]*SkipNode
	maxHeight     int16
}

type SkipNode struct {
	forward [maxNodeHeight]*SkipNode
	Parent  *SkipList
	Level   *enginetypes.PriceLevel
}

func New(listType ListType) *SkipList {
	newSkipList := &SkipList{
		Size:          0,
		TotalQuantity: 0,
		maxHeight:     1,
		nodes:         make(map[int64]*SkipNode),
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

func (this *SkipList) Add(order *enginetypes.EngineOrder) {
	if order == nil {
		return
	}

	if level, ok := this.nodes[order.GetLevel()]; ok {
		level.Level.Add(order)
		return
	}

	newNode := &SkipNode{
		Parent: this,
	}
	newLevel := enginetypes.NewPriceLevel(order.GetLevel(), 0, &doublylinkedlist.DoublyLinkedList{}, newNode)
	newNode.Level = newLevel

	newLevel.Add(order)

	nodeHeight := getRandomHeight()
	if this.maxHeight < nodeHeight {
		this.maxHeight = nodeHeight
	}

	curr := this.Head

	for i := this.maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && this.compare(curr.forward[i].Level.GetLevel(), order.GetLevel()) {
			curr = curr.forward[i]
		}

		if i <= nodeHeight {
			newNode.forward[i] = curr.forward[i]
			curr.forward[i] = newNode
		}
	}

	this.nodes[order.GetLevel()] = newNode
	this.Size++
}

func (this *SkipList) GetRange(quantity int64, price int64) []*enginetypes.PriceLevel {
	var values []*enginetypes.PriceLevel

	if this.Head.forward[0] != nil {
		values = make([]*enginetypes.PriceLevel, 0, 4)
		curr := this.Head.forward[0]
		for quantity > 0 && curr != nil {

			if this.compare(curr.Level.GetLevel(), price) || curr.Level.GetLevel() != price {
				break
			}
			quantity -= curr.Level.GetQuantity()
			values = append(values, curr.Level)

			curr = curr.forward[0]
		}
	}

	return values
}

func (this *SkipList) GetFirst() *enginetypes.PriceLevel {
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
		for curr.forward[i] != nil && this.compare(curr.forward[i].Level.GetLevel(), level) {
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

func getRandomHeight() int16 {
	var height int16 = 1

	for height < maxNodeHeight && rand.Int32N(4) == 0 {
		height++
	}
	return height
}

func (this *SkipNode) Delete() {
	this.Parent.Delete(this.Level.GetLevel())
}
