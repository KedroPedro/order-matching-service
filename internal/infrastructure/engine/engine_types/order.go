package enginetypes

//TODO: rewrite

import (
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	engineinterfaces "github.com/KedroPedro/order-matching-engine/internal/infrastructure/engine/engine_interfaces"
)

type EngineOrder struct {
	order     *entity.Order
	Parent    engineinterfaces.Container
	EnteredAt time.Time
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

func (this EngineOrder) GetLevel() int64 {
	return this.order.Price
}

func (this EngineOrder) GetClass() EngineOrderClass {
	switch this.order.Class {
	case entity.Limit:
		return Limit
	case entity.Market:
		return Market
	case entity.Stop:
		return Stop
	default:
		return UndefinedClass
	}
}

func (this *EngineOrder) Fill(quantity, price int64) (unfilled int64, rest int64) {
	if maxQuantity := this.order.Reserve / price; maxQuantity < quantity {
		quantity = maxQuantity
	}

	diff := this.order.Quantity - this.order.FilledQuantity - quantity
	if diff > 0 {
		this.order.FilledQuantity += quantity
		return diff, 0
	} else if diff < 0 {
		this.order.FilledQuantity += quantity - diff
		return 0, -diff
	} else {
		this.order.FilledQuantity += quantity
		return 0, 0
	}
}

func (this EngineOrder) GetOrderStatus() EngineOrderStatus {
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

//TODO: add event channel

func (this *EngineOrder) SetRejectedStatus() {
	this.order.Status = entity.Rejected
}

func (this *EngineOrder) SetCanceledStatus() {
	this.order.Status = entity.Canceled
}

func (this *EngineOrder) Delete() {
	if this.Parent != nil {
		this.Parent.Delete()
	}
}
