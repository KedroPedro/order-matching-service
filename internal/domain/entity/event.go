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
	Type         EventType
	Order        *Order
	OrderId      string
	Price        int64
	Quantity     int64
	FilledDelta  int64
	NewFilledQty int64
	Delta        int64
	Status       OrderStatus
	OrderType    OrderType
	ReserveDelta int64
}

func NewOrderBeingFilledEvent(order *Order, filledDelta int64, newFilledQty int64, reserveDelta int64) Event {
	return Event{
		Type:         OrderBeingFilled,
		Order:        order,
		FilledDelta:  filledDelta,
		NewFilledQty: newFilledQty,
		ReserveDelta: reserveDelta,
	}
}

func NewOrderQuantityChangedEvent(orderId string, quantityDelta int64, price int64, orderType OrderType) Event {
	return Event{
		Type:      OrderQuantityChanged,
		OrderId:   orderId,
		Quantity:  quantityDelta,
		Price:     price,
		OrderType: orderType,
	}
}

func NewOrderStatusChangedEvent(orderId string, status OrderStatus) Event {
	return Event{
		Type:    OrderStatusChanged,
		OrderId: orderId,
		Status:  status,
	}
}

func NewOrderCancelledEvent(order *Order, delta int64) Event {
	return Event{
		Type:  OrderCancelled,
		Order: order,
		Delta: delta,
	}
}

func NewOrderRejectedEvent(order *Order, delta int64) Event {
	return Event{
		Type:  OrderRejected,
		Order: order,
		Delta: delta,
	}
}

func NewOrderFilledEvent(order *Order, delta int64) Event {
	return Event{
		Type:  OrderFilled,
		Order: order,
		Delta: delta,
	}
}

func NewOrderReserveChangedEvent(orderId string, reserveDelta int64) Event {
	return Event{
		Type:         OrderReserveChanged,
		OrderId:      orderId,
		ReserveDelta: reserveDelta,
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
