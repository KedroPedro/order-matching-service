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
	maxHeight     int16
	compare       func(i, j int64) bool
}

type SkipNode struct {
	forward []*SkipNode
	Parent  *SkipList
	Level   *enginetypes.PriceLevel
}

func New(listType ListType) *SkipList {
	newSkipList := &SkipList{
		Size:          0,
		TotalQuantity: 0,
		maxHeight:     1,
	}

	newSkipList.Head = &SkipNode{
		forward: make([]*SkipNode, maxNodeHeight),
		Parent:  newSkipList,
		Level:   nil,
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

	target, used := this.findNode(order.GetLevel())

	if target == nil || target.Level.GetLevel() != order.GetLevel() {
		newNode := &SkipNode{
			forward: make([]*SkipNode, maxNodeHeight),
			Parent:  this,
		}
		newLevel := enginetypes.NewPriceLevel(order.GetLevel(), 0, &doublylinkedlist.DoublyLinkedList{}, newNode)
		newNode.Level = newLevel

		nodeHeight := getRandomHeight()
		if this.maxHeight < nodeHeight {
			this.maxHeight = nodeHeight
		}

		for i := range this.maxHeight {
			if used[i] == nil {
				used[i] = this.Head
			}
			newNode.forward[i] = used[i].forward[i]
			used[i].forward[i] = newNode

		}

		target = newNode
		this.Size++
	}

	target.Level.Add(order)
	target.Level.IncreaseQuantity(order.GetUnfilledQuantity())
}

func (this *SkipList) GetRange(quantity int64) []*enginetypes.PriceLevel {
	var currQuantity int64 = 0
	values := make([]*enginetypes.PriceLevel, 0)

	if this.Head.forward[0] != nil {
		curr := this.Head.forward[0]
		for currQuantity < quantity && curr != nil {
			currQuantity += curr.Level.GetQuantity()
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
	target, used := this.findNode(level)

	if target == nil || target.Level.GetLevel() != level {
		return
	}

	for i := int16(0); i <= this.maxHeight-1; i++ {
		if used[i].forward[i] == target {
			used[i].forward[i] = target.forward[i]
		} else {
			break
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

func (this *SkipList) findNode(level int64) (*SkipNode, []*SkipNode) {
	used := make([]*SkipNode, maxNodeHeight)
	curr := this.Head

	for i := this.maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && this.compare(curr.forward[i].Level.GetLevel(), level) {
			curr = curr.forward[i]
		}
		used[i] = curr
	}
	return curr.forward[0], used
}

func (this *SkipNode) Delete() {
	this.Parent.Delete(this.Level.GetLevel())
}
