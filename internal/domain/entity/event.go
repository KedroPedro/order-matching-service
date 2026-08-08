package entity

type EventType string

const (
	OrderStatusChanged   EventType = "order_status_changed"
	OrderBeingFilled     EventType = "order_being_filled_event"
	OrderReserveChanged  EventType = "order_reserve_changed"
	OrderCancelled       EventType = "order_cancelled"
	OrderRejected        EventType = "order_rejected"
	OrderFilled          EventType = "order_filled"
	OrderQuantityChanged EventType = "order_quantity_changed"
)

type Event struct {
	payload   any
	eventType EventType
}

type OrderBeingFilledPayload struct {
	Order        *Order
	FilledDelta  int64
	NewFilledQty int64
}

type OrderReserveChangedPayload struct {
	OrderId      string
	ReserveDelta int64
}

type OrderQuantityChangedPayload struct {
	OrderId       string
	QuantityDelta int64
	Price         int64
	OrderType     OrderType
}

type OrderStatusChangedPayload struct {
	OrderId string
	Status  OrderStatus
}

type OrderRemovalPayload struct {
	Order *Order
	Delta int64
}

func NewOrderBeingFilledEvent(order *Order, filledDelta int64, newFilledQty int64) *Event {
	return &Event{
		payload: OrderBeingFilledPayload{
			Order:        order,
			FilledDelta:  filledDelta,
			NewFilledQty: newFilledQty,
		},
		eventType: OrderBeingFilled,
	}
}

func NewOrderReserveChangedEvent(orderId string, reserveDelta int64) *Event {
	return &Event{
		payload: OrderReserveChangedPayload{
			OrderId:      orderId,
			ReserveDelta: reserveDelta,
		},
		eventType: OrderReserveChanged,
	}
}

func NewOrderQuantityChangedEvent(orderId string, quantityDelta int64, price int64, orderType OrderType) *Event {
	return &Event{
		payload: OrderQuantityChangedPayload{
			OrderId:       orderId,
			QuantityDelta: quantityDelta,
			Price:         price,
			OrderType:     orderType,
		},
		eventType: OrderQuantityChanged,
	}
}

func NewOrderStatusChangedEvent(orderId string, status OrderStatus) *Event {
	return &Event{
		payload: OrderStatusChangedPayload{
			OrderId: orderId,
			Status:  status,
		},
		eventType: OrderStatusChanged,
	}
}

func NewOrderCancelledEvent(order *Order, delta int64) *Event {
	return &Event{
		payload:   OrderRemovalPayload{Order: order, Delta: delta},
		eventType: OrderCancelled,
	}
}

func NewOrderRejectedEvent(order *Order, delta int64) *Event {
	return &Event{
		payload:   OrderRemovalPayload{Order: order, Delta: delta},
		eventType: OrderRejected,
	}
}

func NewOrderFilledEvent(order *Order, delta int64) *Event {
	return &Event{
		payload:   OrderRemovalPayload{Order: order, Delta: delta},
		eventType: OrderFilled,
	}
}

func (e *Event) GetType() EventType {
	return e.eventType
}

func (e *Event) GetPayload() any {
	return e.payload
}
