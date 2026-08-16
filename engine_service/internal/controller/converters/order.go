package converters

import (
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/KedroPedro/order-matching-engine/proto"
)

func OrderFromProtoToDomain(protoOrder *proto.Order) (*entity.Order, error) {
	var tif entity.TIF
	switch protoOrder.TimeInForce {
	case proto.TimeInForce_DAY:
		tif = entity.DAY
	case proto.TimeInForce_FOK:
		tif = entity.FOK
	case proto.TimeInForce_GTC:
		tif = entity.GTC
	case proto.TimeInForce_IOC:
		tif = entity.IOC
	default:
		return nil, errs.NewTypeError("invalid time in force")
	}

	var status entity.OrderStatus
	switch protoOrder.Status {
	case proto.Status_Canceled:
		status = entity.Canceled
	case proto.Status_Expired:
		status = entity.Expired
	case proto.Status_Filled:
		status = entity.Filled
	case proto.Status_New:
		status = entity.New
	case proto.Status_PartiallyFilled:
		status = entity.PartiallyFilled
	case proto.Status_Pending:
		status = entity.Pending
	case proto.Status_Rejected:
		status = entity.Rejected
	default:
		return nil, errs.NewTypeError("invalid order status")
	}

	var tp entity.OrderType
	switch protoOrder.Type {
	case proto.Type_Ask:
		tp = entity.Ask
	case proto.Type_Bid:
		tp = entity.Bid
	default:
		return nil, errs.NewTypeError("invalid order type")
	}

	var class entity.OrderClass
	switch protoOrder.Class {
	case proto.Class_Limit:
		class = entity.Limit
	case proto.Class_Market:
		class = entity.Market
	default:
		return nil, errs.NewTypeError("invalid order type")
	}

	order := &entity.Order{
		Id:             protoOrder.Id,
		OwnerId:        protoOrder.OwnerId,
		Type:           tp,
		TimeInForce:    tif,
		Class:          class,
		Status:         status,
		Quantity:       protoOrder.Quantity,
		FilledQuantity: protoOrder.FilledQuantity,
		Price:          protoOrder.Price,
		Reserve:        protoOrder.Reserve,
		CreatedAt:      protoOrder.CreatedAt.AsTime(),
	}
	order.SetStopStatus(protoOrder.Stop)

	return order, nil
}
