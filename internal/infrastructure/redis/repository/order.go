package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/redis/go-redis/v9"
)

const (
	incrLevelScript = `
		local newQty = redis.call('HINCRBY', KEYS[1], ARGV[1], ARGV[2])
		if newQty <= 0 then
    		redis.call('HDEL', KEYS[1], ARGV[1])
		end
		return newQty
	`
)

type OrderRepository struct {
	conn         *redis.Client
	incrLevelSha string
}

func NewOrderRepository(conn *redis.Client) *OrderRepository {
	return &OrderRepository{
		conn: conn,
	}
}

func (this *OrderRepository) LoadScripts(ctx context.Context) error {
	sha, err := this.conn.ScriptLoad(ctx, incrLevelScript).Result()
	if err != nil {
		return fmt.Errorf("load incr_level.lua: %w", err)
	}
	this.incrLevelSha = sha
	return nil
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
			"type":            string(order.Type),
			"time_in_force":   string(order.TimeInForce),
			"price":           order.Price,
			"stop":            order.Stop,
			"filled_quantity": order.FilledQuantity,
			"class":           string(order.Class),
			"quantity":        order.Quantity,
			"status":          string(order.Status),
		}),
	}

	pipe.EvalSha(
		ctx,
		this.incrLevelSha,
		[]string{fmt.Sprintf("%s:%s", "book", order.Type)},
		strconv.Itoa(int(order.Price)),
		order.Quantity,
	)

	if _, err := pipe.Exec(ctx); err != nil {
		return errs.NewRepositoryError("redis pipe execution error", err)
	}

	for _, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			return errs.NewRepositoryError("redis request error", err)
		}
	}

	return nil
}

func (this *OrderRepository) ProcessEvent(ctx context.Context, event *entity.Event) error {
	var err error

	switch event.GetType() {
	case entity.OrderBeingFilled:
		payload := event.GetPayload().(entity.OrderBeingFilledPayload)
		orderKey := fmt.Sprintf("order:%s", payload.Order.Id)
		bookKey := fmt.Sprintf("book:%s", string(payload.Order.Type))

		pipe := this.conn.Pipeline()
		pipe.HSet(ctx, orderKey, "filled_quantity", payload.NewFilledQty)
		pipe.EvalSha(
			ctx,
			this.incrLevelSha,
			[]string{bookKey},
			strconv.Itoa(int(payload.Order.Price)),
			-payload.FilledDelta,
		)

		_, err = pipe.Exec(ctx)

	case entity.OrderReserveChanged:
		payload := event.GetPayload().(entity.OrderReserveChangedPayload)
		orderKey := fmt.Sprintf("order:%s", payload.OrderId)
		err = this.conn.HIncrBy(ctx, orderKey, "reserve", -payload.ReserveDelta).Err()

	case entity.OrderQuantityChanged:
		payload := event.GetPayload().(entity.OrderQuantityChangedPayload)
		bookKey := fmt.Sprintf("book:%s", string(payload.OrderType))

		err = this.conn.EvalSha(
			ctx,
			this.incrLevelSha,
			[]string{bookKey},
			strconv.Itoa(int(payload.Price)),
			-payload.QuantityDelta,
		).Err()

	case entity.OrderStatusChanged:
		payload := event.GetPayload().(entity.OrderStatusChangedPayload)
		orderKey := fmt.Sprintf("order:%s", payload.OrderId)
		err = this.conn.HSet(ctx, orderKey, "status", string(payload.Status)).Err()

	case entity.OrderCancelled, entity.OrderRejected, entity.OrderFilled:
		payload := event.GetPayload().(entity.OrderRemovalPayload)
		orderKey := fmt.Sprintf("order:%s", payload.Order.Id)
		bookKey := fmt.Sprintf("book:%s", string(payload.Order.Type))
		orderBookMember := fmt.Sprintf("%d:%s", payload.Order.CreatedAt.UnixNano(), payload.Order.Id)

		pipe := this.conn.Pipeline()
		pipe.Del(ctx, orderKey)
		pipe.ZRem(ctx, string(payload.Order.Type), orderBookMember)
		pipe.EvalSha(
			ctx,
			this.incrLevelSha,
			[]string{bookKey},
			strconv.Itoa(int(payload.Order.Price)),
			-payload.Delta,
		)
		_, err = pipe.Exec(ctx)
	}

	if err != nil {
		return errs.NewRepositoryError("event processing error", err)
	}

	return nil
}

func (this *OrderRepository) GetById(ctx context.Context, orderId string) (*entity.Order, error) {
	fields, err := this.conn.HGetAll(ctx, fmt.Sprintf("%s:%s", "order", orderId)).Result()
	if err != nil {
		return nil, errs.NewRepositoryError("redis query execution error", err)
	}

	var order entity.Order
	order.Id = fields["id"]
	order.OwnerId = fields["owner_id"]
	order.Type = entity.OrderType(fields["type"])
	order.TimeInForce = entity.OrderTimeInForce(fields["time_in_force"])
	order.Price, _ = strconv.ParseInt(fields["price"], 10, 64)
	order.Stop, _ = strconv.ParseBool(fields["stop"])
	order.FilledQuantity, _ = strconv.ParseInt(fields["filled_quantity"], 10, 64)
	order.Class = entity.OrderClass(fields["class"])
	order.Quantity, _ = strconv.ParseInt(fields["quantity"], 10, 64)
	order.Status = entity.OrderStatus(fields["status"])
	order.Reserve, _ = strconv.ParseInt(fields["reserve"], 10, 64)

	return &order, nil
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
		return -1, errs.NewRepositoryError("redis query execution error", err)
	}

	if len(z) == 0 {
		return -1, errs.NewRepositoryError("no values in storage", errors.New("empty storage"))
	}

	return int64(z[0].Score), nil
}

func (this *OrderRepository) GetState(ctx context.Context) (asks map[string]string, bids map[string]string, err error) {

	pipe := this.conn.Pipeline()

	mapCmds := []*redis.MapStringStringCmd{
		pipe.HGetAll(ctx, fmt.Sprintf("%s:%s", "book", string(entity.Ask))),
		pipe.HGetAll(ctx, fmt.Sprintf("%s:%s", "book", string(entity.Bid))),
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, nil, errs.NewRepositoryError("redis pipe execution error", err)
	}

	maps := make([]map[string]string, len(mapCmds))
	for i, cmd := range mapCmds {
		if m, err := cmd.Result(); err != nil {
			return nil, nil, errs.NewRepositoryError("redis query execution error", err)
		} else {
			maps[i] = m
		}
	}

	return maps[0], maps[1], nil
}
