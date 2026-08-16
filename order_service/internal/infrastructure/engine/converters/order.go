package converters

import (
	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/KedroPedro/order-matching-engine/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func OrderFromDomainToProto(order *entity.Order) (*proto.Order, error) {
	var tif proto.TimeInForce
	switch order.TimeInForce {
	case entity.Day:
		tif = proto.TimeInForce_DAY
	case entity.Fok:
		tif = proto.TimeInForce_FOK
	case entity.Gtc:
		tif = proto.TimeInForce_GTC
	case entity.Ioc:
		tif = proto.TimeInForce_IOC
	default:
		return nil, errs.NewTypeError("invalid time in force")
	}

	var status proto.Status
	switch order.Status {
	case entity.Canceled:
		status = proto.Status_Canceled
	case entity.Expired:
		status = proto.Status_Expired
	case entity.Filled:
		status = proto.Status_Filled
	case entity.New:
		status = proto.Status_New
	case entity.PartiallyFilled:
		status = proto.Status_PartiallyFilled
	case entity.Pending:
		status = proto.Status_Pending
	case entity.Rejected:
		status = proto.Status_Rejected
	default:
		return nil, errs.NewTypeError("invalid order status")
	}

	var tp proto.Type
	switch order.Type {
	case entity.Ask:
		tp = proto.Type_Ask
	case entity.Bid:
		tp = proto.Type_Bid
	default:
		return nil, errs.NewTypeError("invalid order type")
	}

	var class proto.Class
	switch order.Class {
	case entity.Limit:
		class = proto.Class_Limit
	case entity.Market:
		class = proto.Class_Market
	default:
		return nil, errs.NewTypeError("invalid order type")
	}

	return &proto.Order{
		Id:             order.Id,
		OwnerId:        order.OwnerId,
		Type:           tp,
		TimeInForce:    tif,
		Class:          class,
		Status:         status,
		Quantity:       order.Quantity,
		FilledQuantity: order.FilledQuantity,
		Price:          order.Price,
		Reserve:        order.Reserve,
		CreatedAt:      timestamppb.New(order.CreatedAt),
		Stop:           order.Stop,
	}, nil

}
