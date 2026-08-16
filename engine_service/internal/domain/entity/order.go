package entity

import "time"

type Container interface {
	Delete()
}

type Order struct {
	Id             string
	OwnerId        string
	Type           OrderType
	TimeInForce    TIF
	Class          OrderClass
	Status         OrderStatus
	Quantity       int64
	FilledQuantity int64
	Price          int64
	Reserve        int64
	CreatedAt      time.Time
	stop           bool
}

type OrderType string

const (
	Bid OrderType = "bid"
	Ask OrderType = "ask"
)

type TIF string

const (
	FOK TIF = "fok"
	IOC TIF = "ioc"
	DAY TIF = "day"
	GTC TIF = "gtc"
)

type OrderClass string

const (
	Market OrderClass = "market"
	Limit  OrderClass = "limit"
	Stop   OrderClass = "stop"
)

type OrderStatus string

const (
	New             OrderStatus = "new"
	Pending         OrderStatus = "pending"
	PartiallyFilled OrderStatus = "partially_filled"
	Filled          OrderStatus = "filled"
	Canceled        OrderStatus = "canceled"
	Rejected        OrderStatus = "rejected"
	Expired         OrderStatus = "expired"
	UndefinedStatus OrderStatus = "undefined_status"
)

func (this Order) GetUnfilledQuantity() int64 {
	return this.Quantity - this.FilledQuantity
}

func (this Order) GetFilledQuantity() int64 {
	return this.FilledQuantity
}

func (this *Order) ActivateStopOrder() {
	this.stop = false
}

func (this Order) IsStopOrder() bool {
	return this.stop
}

func (this *Order) SetStopStatus(isStop bool) {
	this.stop = isStop
}

func (this *Order) Fill(quantity, price int64) (event Event) {
	this.FilledQuantity += quantity
	this.Reserve -= quantity * price

	return NewOrderBeingFilledEvent(this, quantity, this.FilledQuantity, quantity*price)
}

func (this *Order) SetNewStatus() Event {
	this.Status = New
	return NewOrderStatusChangedEvent(this.Id, New)
}

func (this *Order) SetPendingStatus() Event {
	this.Status = Pending
	return NewOrderStatusChangedEvent(this.Id, Pending)
}

func (this *Order) SetExpiredStatus() Event {
	this.Status = Expired
	return NewOrderStatusChangedEvent(this.Id, Expired)
}

func (this *Order) SetRejectedStatus() Event {
	this.Status = Rejected
	return NewOrderRejectedEvent(this)
}

func (this *Order) SetCanceledStatus() Event {
	this.Status = Canceled
	return NewOrderCancelledEvent(this)
}

func (this *Order) SetPartiallyFilledStatus() Event {
	this.Status = PartiallyFilled
	return NewOrderStatusChangedEvent(this.Id, PartiallyFilled)
}

func (this *Order) SetFilledStatus() Event {
	this.Status = Filled
	return NewOrderFilledEvent(this)
}
