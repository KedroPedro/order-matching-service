package entity

import "sync"

type EventType string

const (
	OrderStatusChanged EventType = "order_status_changed"
	OrderBeingFilled   EventType = "order_being_filled_event"
	OrderCancelled     EventType = "order_cancelled"
	OrderRejected      EventType = "order_rejected"
	OrderFilled        EventType = "order_filled"
)

type Event struct {
	Type         EventType
	Order        *Order //TODO: delete maybe
	OrderId      string
	Price        int64
	Quantity     int64
	FilledDelta  int64
	NewFilledQty int64
	Delta        int64
	Status       OrderStatus //TODO: also delete maybe
	ReserveDelta int64
}

func NewOrderBeingFilledEvent(order *Order, filledDelta int64, newFilledQty int64, reserveDelta int64) Event {
	return Event{
		Type:         OrderBeingFilled,
		Order:        order,
		Status:       order.Status,
		FilledDelta:  filledDelta,
		NewFilledQty: newFilledQty,
		ReserveDelta: reserveDelta,
	}
}

func NewOrderStatusChangedEvent(orderId string, status OrderStatus) Event {
	return Event{
		Type:    OrderStatusChanged,
		OrderId: orderId,
		Status:  status,
	}
}

func NewOrderCancelledEvent(order *Order) Event {
	return Event{
		Type:   OrderCancelled,
		Status: order.Status,
		Order:  order,
		Delta:  order.GetUnfilledQuantity(),
	}
}

func NewOrderRejectedEvent(order *Order) Event {
	return Event{
		Type:   OrderRejected,
		Status: order.Status,
		Order:  order,
		Delta:  order.Quantity,
	}
}

func NewOrderFilledEvent(order *Order) Event {
	return Event{
		Type:   OrderFilled,
		Status: order.Status,
		Order:  order,
		Delta:  order.GetUnfilledQuantity(),
	}
}

func (e Event) GetType() EventType {
	return e.Type
}

func (e Event) GetOrder() *Order {
	return e.Order
}

func (e Event) GetOrderId() string {
	if e.Order != nil {
		return e.Order.Id
	}
	return e.OrderId
}

//================================================================

const (
	eventBatchSize = 1024
)

var eventBatchPool = sync.Pool{
	New: func() any {
		return &EventBatch{
			events: make([]Event, 0, eventBatchSize),
		}
	},
}

type EventBatch struct {
	events []Event
}

func NewEventBatch() *EventBatch {
	batch := eventBatchPool.Get().(*EventBatch)
	batch.Reset()
	return batch
}

func (this *EventBatch) Add(event Event) {
	this.events = append(this.events, event)
}

func (this *EventBatch) GetEvents() []Event {
	return this.events
}

func (this *EventBatch) Reset() {
	this.events = this.events[:0]
}

func (this *EventBatch) Len() int {
	return len(this.events)
}

func (this *EventBatch) Release() {
	eventBatchPool.Put(this)
}
