package repository

import (
	"context"
	"fmt"
	"strconv"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/redis/go-redis/v9"
)

type OrderRepository struct {
	conn *redis.Conn
}

func NewOrderRepository(conn *redis.Conn) *OrderRepository {
	return &OrderRepository{
		conn: conn,
	}
}

func (this *OrderRepository) AddToQueue(ctx context.Context, order *entity.Order) error {
	pipe := this.conn.Pipeline()

	cmds := []*redis.IntCmd{
		pipe.ZAdd(
			ctx,
			string(order.Type),
			redis.Z{
				Score:  float64(order.Price),
				Member: fmt.Sprintf("%d:%s", order.CreatedAt.UnixNano(), order.Id),
			},
		),
		pipe.HSet(ctx, fmt.Sprintf("%s:%s", "order", order.Id), map[string]interface{}{
			"id":              order.Id,
			"owner_id":        order.OwnerId,
			"type":            order.Type,
			"time_in_force":   order.TimeInForce,
			"price":           order.Price,
			"stop":            order.Stop,
			"filled_quantity": order.FilledQuantity,
			"class":           order.Class,
			"product_id":      order.ProductId,
			"quantity":        order.Quantity,
			"status":          order.Status,
		}),
		pipe.HIncrBy(ctx, fmt.Sprintf("%s:%s", "book", order.Type), strconv.Itoa(int(order.Price)), order.Quantity),
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			return err //TODO: fix
		}
	}

	return nil
}

func (this *OrderRepository) ProcessEvent(ctx context.Context, event *entity.Event) error {
	var err error

	switch event.GetType() {
	case entity.OrderBeingFilled:
		pipe := this.conn.Pipeline()
		pipe.HIncrBy(ctx, fmt.Sprintf("%s:%s", "order", event.GetOrderId()), "filled_quantity", event.GetValue().(int64))
		pipe.HIncrBy(ctx, fmt.Sprintf("%s:%s", "book", event.GetOrderType()), strconv.Itoa(int(event.GetOrderPrice())), -event.GetValue().(int64))

	case entity.OrderReserveChanged:
		err = this.conn.HIncrBy(ctx, fmt.Sprintf("%s:%s", "order", event.GetOrderId()), "reserve", -event.GetValue().(int64)).Err()

	default:
		err = this.conn.HSet(ctx, fmt.Sprintf("%s:%s", "order", event.GetOrderId()), "status", event.GetValue().(string)).Err()
	}

	return err
}

func (this *OrderRepository) GetBestPrice(ctx context.Context, orderType entity.OrderType) (int64, error) {
	var slice *redis.ZSliceCmd

	switch orderType {
	case entity.Ask:
		slice = this.conn.ZRevRangeWithScores(ctx, string(entity.Bid), 0, 0)

	case entity.Bid:
		slice = this.conn.ZRangeWithScores(ctx, string(entity.Ask), 0, 0)
	}

	z, err := slice.Result()
	if err != nil {
		return -1, err
	}

	if len(z) == 0 {
		return -1, nil
	}

	return int64(z[0].Score), nil
}

func (this *OrderRepository) GetState(ctx context.Context) (asks map[string]string, bids map[string]string, err error) {

	pipe := this.conn.Pipeline()

	mapCmds := []*redis.MapStringStringCmd{
		pipe.HGetAll(ctx, fmt.Sprintf("%s:%s", "book", entity.Ask)),
		pipe.HGetAll(ctx, fmt.Sprintf("%s:%s", "book", entity.Bid)),
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, nil, err
	}

	maps := make([]map[string]string, len(mapCmds))
	for i, cmd := range mapCmds {
		if m, err := cmd.Result(); err != nil {
			return nil, nil, err
		} else {
			maps[i] = m
		}
	}

	return maps[0], maps[1], nil
}
