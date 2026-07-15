package skiplist

import (
	"math/rand/v2"

	enginetypes "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_types"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/storages/doublylinkedlist"
)

const (
	maxNodeHeight = 16
)

type SkipList struct {
	Head      *SkipNode
	Size      int64
	maxHeight int16
}

type SkipNode struct {
	forward []*SkipNode
	Parent  *SkipList
	Value   *doublylinkedlist.DoublyLinkedList
	Level   int64
}

func (this *SkipList) Add(order *enginetypes.EngineOrder) {
	curr, update := findNode(this.Head, order.GetLevel(), this.maxHeight)

	for i := this.maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && curr.forward[i].Level <= order.GetLevel() {
			curr = curr.forward[i]
		}
		update = append(update, curr)
	}

	if curr.Level != order.GetLevel() {
		newNode := SkipNode{
			Parent:  this,
			forward: make([]*SkipNode, 0, maxNodeHeight),
			Level:   order.GetLevel(),
			Value:   &doublylinkedlist.DoublyLinkedList{Level: order.GetLevel()},
		}
		newNode.Value.Parent = &newNode

		nodeHeight := getRandomHeight()

		if nodeHeight > this.maxHeight {
			this.maxHeight = nodeHeight
		}

		for i := len(update) - 1; i >= 0; i++ {
			uNode := update[i]
			for j := range nodeHeight {
				if (uNode.forward[j] == nil || uNode.forward[j].Level > newNode.Level) && j <= nodeHeight {
					newNode.forward[j] = uNode.forward[j]
					uNode.forward[j] = &newNode
				} else {
					break
				}
			}
		}
		curr = &newNode
	}
	this.Size++
	curr.Value.Add(order)
}

func (this *SkipList) GetRange(size int64) []*doublylinkedlist.DoublyLinkedList {
	var currSize int64 = 0
	values := make([]*doublylinkedlist.DoublyLinkedList, 0)

	for currSize < size {
		curr := this.Head.forward[0]

		currSize += curr.Value.Size
		values = append(values, curr.Value)

		curr = curr.forward[0]
	}

	return values
}

func (this *SkipList) GetFirst() *doublylinkedlist.DoublyLinkedList {
	return this.Head.forward[0].Value
}

func (this *SkipList) Get(level int64) *doublylinkedlist.DoublyLinkedList {
	curr, _ := findNode(this.Head, level, this.maxHeight)

	if curr.Level != level {
		return nil
	}

	return curr.Value
}

func (this *SkipList) Delete(level int64) {
	curr, update := findNode(this.Head, level, this.maxHeight)

	if curr.Level != level {
		return
	}

	for i := len(update) - 1; i >= 0; i-- {
		uNode := update[i]
		for j := range this.maxHeight {
			if uNode.forward[j] != nil && uNode.forward[j] == curr {
				uNode.forward[j] = curr.forward[j]
			}
		}
	}
	this.Size--
}

func (this *SkipList) DeleteFirst() {
	firstNode := this.Head.forward[0]

	for currHeight, currNode := range this.Head.forward {
		if currNode != firstNode {
			break
		}

		this.Head.forward[currHeight] = currNode.forward[currHeight]
	}
	this.Size--
}

func getRandomHeight() int16 {
	var height int16 = 1
	for i := 0; i < maxNodeHeight && rand.Int32N(4) == 0; i++ {
		height++
	}
	return height
}

func findNode(start *SkipNode, level int64, maxHeight int16) (curr *SkipNode, used []*SkipNode) {
	update := make([]*SkipNode, 0, maxNodeHeight)
	curr = start

	for i := maxHeight - 1; i >= 0; i-- {
		for curr.forward[i] != nil && curr.forward[i].Level <= level {
			curr = curr.forward[i]
		}
		update = append(update, curr)
	}

	return curr, used
}

func (this *SkipNode) Delete() {
	this.Parent.Delete(this.Level)
}
