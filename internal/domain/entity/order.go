package entity

import "time"

type Order struct {
	Id             string
	OwnerId        string
	Type           OrderType
	Price          int64
	Quantity       int64
	FilledQuantity int64
	Class          OrderClass
	TimeInForce    OrderTimeInForce
	Status         OrderStatus
	ProductId      string
	CreatedAt      time.Time
	ClosedAt       time.Time
	Reserve        int64
}

type OrderType string

const (
	Ask OrderType = "ask"
	Bid OrderType = "bid"
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
)

type OrderTimeInForce string

const (
	Day OrderTimeInForce = "day"
	Gtc OrderTimeInForce = "gtc"
	Ioc OrderTimeInForce = "ioc"
	Fok OrderTimeInForce = "fok"
)
