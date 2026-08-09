package enginetypes

import (
	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
)

type Container interface {
	Delete()
}

type EngineOrder struct {
	order  *entity.Order
	Parent Container
}

func NewEngineOrder(
	order *entity.Order,
	parent Container,
) *EngineOrder {
	return &EngineOrder{
		order:  order,
		Parent: parent,
	}
}

type EngineOrderType string

const (
	Bid           EngineOrderType = "bid"
	Ask           EngineOrderType = "ask"
	UndefinedType                 = "undefined_type"
)

type TIF string

const (
	FOK          TIF = "fok"
	IOC          TIF = "ioc"
	DAY          TIF = "day"
	GTC          TIF = "gtc"
	UndefinedTIF     = "undefined_tif"
)

type EngineOrderClass string

const (
	Market         EngineOrderClass = "market"
	Limit          EngineOrderClass = "limit"
	Stop           EngineOrderClass = "stop"
	UndefinedClass EngineOrderClass = "undefined"
)

type EngineOrderStatus string

const (
	New             EngineOrderStatus = "new"
	Pending         EngineOrderStatus = "pending"
	PartiallyFilled EngineOrderStatus = "partially_filled"
	Filled          EngineOrderStatus = "filled"
	Canceled        EngineOrderStatus = "canceled"
	Rejected        EngineOrderStatus = "rejected"
	Expired         EngineOrderStatus = "expired"
	UndefinedStatus EngineOrderStatus = "undefined_status"
)

func (this EngineOrder) GetType() EngineOrderType {
	switch this.order.Type {
	case entity.Ask:
		return Ask
	case entity.Bid:
		return Bid
	default:
		return UndefinedType
	}
}

func (this EngineOrder) GetTimeInForce() TIF {
	switch this.order.TimeInForce {
	case entity.Day:
		return DAY
	case entity.Fok:
		return FOK
	case entity.Gtc:
		return GTC
	case entity.Ioc:
		return IOC
	default:
		return UndefinedTIF
	}
}

func (this EngineOrder) GetId() string {
	return this.order.Id
}

func (this EngineOrder) GetQuantity() int64 {
	return this.order.Quantity
}

func (this EngineOrder) GetUnfilledQuantity() int64 {
	return this.order.Quantity - this.order.FilledQuantity
}

func (this EngineOrder) GetFilledQuantity() int64 {
	return this.order.FilledQuantity
}

func (this EngineOrder) GetLevel() int64 {
	return this.order.Price
}

func (this *EngineOrder) ActivateStopOrder() {
	this.order.Stop = false
}

func (this EngineOrder) IsStopOrder() bool {
	return this.order.Stop
}

func (this EngineOrder) GetClass() EngineOrderClass {
	switch this.order.Class {
	case entity.Limit:
		return Limit
	case entity.Market:
		return Market
	default:
		return UndefinedClass
	}
}

func (this EngineOrder) GetReserve() int64 {
	return this.order.Reserve
}

func (this *EngineOrder) Fill(quantity, price int64) (unfilled int64, rest int64, events []entity.Event) {
	events = make([]entity.Event, 0, 4)

	requested := quantity

	if maxQuantity := this.order.Reserve / price; maxQuantity < quantity {
		quantity = maxQuantity
	}

	events = append(events, this.SetPartiallyFilledStatus())

	diff := this.order.Quantity - this.order.FilledQuantity - quantity
	if diff > 0 {
		this.order.FilledQuantity += quantity
		this.order.Reserve -= quantity * price
		events = append(events, this.fillOrderEvent(quantity, price)...)
		return diff, requested - quantity, events
	} else if diff < 0 {
		remaining := this.order.Quantity - this.order.FilledQuantity
		this.order.FilledQuantity = this.order.Quantity
		this.order.Reserve -= remaining * price
		events = append(events, this.SetFilledStatus())
		events = append(events, this.fillOrderEvent(remaining, price)...)
		return 0, requested - remaining, events
	} else {
		this.order.FilledQuantity += quantity
		this.order.Reserve -= quantity * price
		events = append(events, this.fillOrderEvent(quantity, price)...)
		events = append(events, this.SetFilledStatus())
		return 0, requested - quantity, events
	}
}

func (this EngineOrder) GetStatus() EngineOrderStatus {
	switch this.order.Status {
	case entity.Canceled:
		return Canceled
	case entity.Expired:
		return Expired
	case entity.Filled:
		return Filled
	case entity.New:
		return New
	case entity.PartiallyFilled:
		return PartiallyFilled
	case entity.Pending:
		return Pending
	case entity.Rejected:
		return Rejected
	default:
		return UndefinedStatus
	}
}

func (this *EngineOrder) SetNewStatus() entity.Event {
	this.order.Status = entity.New
	return entity.NewOrderStatusChangedEvent(this.order.Id, entity.New)
}

func (this *EngineOrder) SetPendingStatus() entity.Event {
	this.order.Status = entity.Pending
	return entity.NewOrderStatusChangedEvent(this.order.Id, entity.Pending)
}

func (this *EngineOrder) SetExpiredStatus() entity.Event {
	this.order.Status = entity.Expired
	return entity.NewOrderStatusChangedEvent(this.order.Id, entity.Expired)
}

func (this *EngineOrder) SetRejectedStatus() entity.Event {
	this.order.Status = entity.Rejected
	return entity.NewOrderRejectedEvent(this.order, this.order.Quantity)
}

func (this *EngineOrder) SetCanceledStatus() entity.Event {
	this.order.Status = entity.Canceled
	return entity.NewOrderCancelledEvent(this.order, this.order.Quantity-this.order.FilledQuantity)
}

func (this *EngineOrder) SetPartiallyFilledStatus() entity.Event {
	this.order.Status = entity.PartiallyFilled
	return entity.NewOrderStatusChangedEvent(this.order.Id, entity.PartiallyFilled)
}

func (this *EngineOrder) SetFilledStatus() entity.Event {
	this.order.Status = entity.Filled
	return entity.NewOrderFilledEvent(this.order, this.order.Quantity-this.order.FilledQuantity)
}

func (this *EngineOrder) fillOrderEvent(quantity, price int64) []entity.Event {
	return []entity.Event{
		entity.NewOrderBeingFilledEvent(this.order, quantity, this.order.FilledQuantity),
		entity.NewOrderReserveChangedEvent(this.order.Id, quantity*price),
	}
}

func (this *EngineOrder) Delete() {
	this.Parent.Delete()
}
