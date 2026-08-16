package types

import (
	"time"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
)

type Order struct {
	Id             string    `bson:"_id" `
	OwnerId        string    `bson:"owner_id" `
	Type           string    `bson:"type" `
	Price          int64     `bson:"price" `
	Quantity       int64     `bson:"quantity"`
	FilledQuantity int64     `bson:"filled_quantity"`
	Class          string    `bson:"class" `
	TimeInForce    string    `bson:"time_in_force"`
	Status         string    `bson:"status" `
	CreatedAt      time.Time `bson:"created_at"`
	Reserve        int64     `bson:"reserve"`
	Stop           bool      `bson:"stop"`
}

func OrderFromDomain(order *entity.Order) *Order {
	return &Order{
		Id:             order.Id,
		OwnerId:        order.OwnerId,
		Type:           string(order.Type),
		Price:          order.Price,
		Quantity:       order.Quantity,
		FilledQuantity: order.FilledQuantity,
		Class:          string(order.Class),
		TimeInForce:    string(order.TimeInForce),
		Status:         string(order.Status),
		CreatedAt:      order.CreatedAt,
		Reserve:        order.Reserve,
		Stop:           order.IsStopOrder(),
	}
}

func OrderToDomain(order *Order) (*entity.Order, error) {
	var Type entity.OrderType
	switch order.Type {
	case string(entity.Ask):
		Type = entity.Ask
	case string(entity.Bid):
		Type = entity.Bid
	default:
		return nil, errs.NewTypeError("undefined order type")
	}

	var Class entity.OrderClass
	switch order.Class {
	case string(entity.Market):
		Class = entity.Market
	case string(entity.Limit):
		Class = entity.Limit
	default:
		return nil, errs.NewTypeError("undefined order class")
	}

	var TimeInForce entity.TIF
	switch order.TimeInForce {
	case string(entity.FOK):
		TimeInForce = entity.FOK
	case string(entity.IOC):
		TimeInForce = entity.IOC
	case string(entity.GTC):
		TimeInForce = entity.GTC
	case string(entity.DAY):
		TimeInForce = entity.DAY
	default:
		return nil, errs.NewTypeError("undefined order time in force")
	}

	var Status entity.OrderStatus
	switch order.Status {
	case string(entity.New):
		Status = entity.New
	case string(entity.Pending):
		Status = entity.Pending
	case string(entity.PartiallyFilled):
		Status = entity.PartiallyFilled
	case string(entity.Filled):
		Status = entity.Filled
	case string(entity.Canceled):
		Status = entity.Canceled
	case string(entity.Rejected):
		Status = entity.Rejected
	case string(entity.Expired):
		Status = entity.Expired
	default:
		return nil, errs.NewTypeError("undefined order status")
	}

	dOrder := &entity.Order{
		Id:             order.Id,
		OwnerId:        order.OwnerId,
		Type:           Type,
		Price:          order.Price,
		Quantity:       order.Quantity,
		FilledQuantity: order.FilledQuantity,
		Class:          Class,
		TimeInForce:    TimeInForce,
		Status:         Status,
		CreatedAt:      order.CreatedAt,
		Reserve:        order.Reserve,
	}
	dOrder.SetStopStatus(order.Stop)

	return dOrder, nil
}
