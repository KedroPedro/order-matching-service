package controllertypes

import (
	"time"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/google/uuid"
)

type Order struct {
	OwnerId     string  `json:"owner_id"`
	Type        string  `json:"type"`
	Price       float64 `json:"price"`
	Quantity    int64   `json:"quantity"`
	Class       string  `json:"class"`
	TimeInForce string  `json:"time_in_force"`
	Stop        bool    `json:"stop"`
}

func (this *Order) ToDomainEntity(ownerId string) (*entity.Order, error) {
	var Type entity.OrderType
	switch this.Type {
	case string(entity.Ask):
		Type = entity.Ask
	case string(entity.Bid):
		Type = entity.Bid
	default:
		return nil, errs.NewTypeError("undefined order type")
	}

	var Class entity.OrderClass
	switch this.Class {
	case string(entity.Market):
		Class = entity.Market
	case string(entity.Limit):
		Class = entity.Limit
	default:
		return nil, errs.NewTypeError("undefined order class")
	}

	var TimeInForce entity.OrderTimeInForce
	switch this.TimeInForce {
	case string(entity.Fok):
		TimeInForce = entity.Fok
	case string(entity.Ioc):
		TimeInForce = entity.Ioc
	case string(entity.Gtc):
		TimeInForce = entity.Gtc
	case string(entity.Day):
		TimeInForce = entity.Day
	default:
		return nil, errs.NewTypeError("undefined order time in force")
	}

	return &entity.Order{
		Id:             uuid.New().String(),
		OwnerId:        ownerId,
		Type:           Type,
		Price:          int64(this.Price * 100),
		Quantity:       this.Quantity,
		FilledQuantity: 0,
		Class:          Class,
		TimeInForce:    TimeInForce,
		Status:         entity.Pending,
		CreatedAt:      time.Now(),
		Stop:           this.Stop,
	}, nil
}
