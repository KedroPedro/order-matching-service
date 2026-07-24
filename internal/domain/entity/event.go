package entity

type Event struct {
	orderId      string
	orderOwnerId string
	orderType    OrderType
	orderPrice   int64
	value        any
	event        EventType
}

type EventType string

const (
	OrderCancelled      EventType = "order_cancelled_event"
	OrderFilled         EventType = "order_filled_event"
	OrderRejected       EventType = "order_rejected_event"
	OrderExpired        EventType = "order_expired_event"
	OrderPending        EventType = "order_pending_event"
	OrderNew            EventType = "order_new_event"
	OrderBeingFilled    EventType = "order_being_filled_event"
	OrderReserveChanged EventType = "order_reserve_changed_event"
)

func NewEvent(order *Order, value any, eventType EventType) *Event {
	return &Event{
		orderId:      order.Id,
		orderOwnerId: order.OwnerId,
		orderType:    order.Type,
		orderPrice:   order.Price,
		value:        value,
		event:        eventType,
	}
}

func (this Event) GetType() EventType {
	return this.event
}

func (this Event) GetValue() any {
	return this.value
}

func (this Event) GetOrderId() string {
	return this.orderId
}

func (this Event) GetOrderType() OrderType {
	return this.orderType
}

func (this Event) GetOrderOwnerId() string {
	return this.orderOwnerId
}

func (this Event) GetOrderPrice() int64 {
	return this.orderPrice
}
