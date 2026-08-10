package eventbatch

import (
	"sync"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
)

const (
	eventBatchSize = 1024
)

var batchPool = sync.Pool{
	New: func() any {
		return &EventBatch{
			events: make([]entity.Event, 0, eventBatchSize),
		}
	},
}

type EventBatch struct {
	events []entity.Event
}

func New() *EventBatch {
	batch := batchPool.Get().(*EventBatch)
	batch.Reset()
	return batch
}

func (this *EventBatch) Add(event entity.Event) {
	this.events = append(this.events, event)
}

func (this *EventBatch) AddMany(events ...entity.Event) {
	this.events = append(this.events, events...)
}

func (this *EventBatch) GetEvents() []entity.Event {
	return this.events
}

func (this *EventBatch) Reset() {
	this.events = this.events[:0]
}

func (this *EventBatch) Release() {
	batchPool.Put(this)
}

func (this *EventBatch) Len() int {
	return len(this.events)
}
