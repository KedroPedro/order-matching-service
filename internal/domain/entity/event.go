package entity

type Event struct {
	orderId      string
	orderOwnerId string
	orderType    OrderType
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
