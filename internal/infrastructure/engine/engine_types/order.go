package enginetypes

import "github.com/KedroPedro/order-matching-engine/internal/domain/entity"

type EngineOrder struct {
	order  *entity.Order
	Parent any
}

type EngineOrderType string

const (
	Bid EngineOrderType = "bid"
	Ask EngineOrderType = "ask"
)

type TIF string

const (
	FOK TIF = "fok"
	IOC TIF = "ioc"
	DAY TIF = "day"
	GTC TIF = "gtc"
)

func (this *EngineOrder) GetLevel() int64 {
	return this.order.Price
}

func (this *EngineOrder) GetType() TIF {
	var tif TIF

	switch this.order.TimeInForce {
	case entity.Day:
		tif = DAY
	case entity.Fok:
		tif = FOK
	case entity.Gtc:
		tif = GTC
	case entity.Ioc:
		tif = IOC
	}

	return tif
}

func (this *EngineOrder) GetFilledQuantity() int64 {
	return this.order.FilledQuantity
}

func (this *EngineOrder) Fill(quantity int64) (lack int64, rem int64) {
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
