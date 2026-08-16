package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
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

func (this *OrderRepository) ProcessEvent(ctx context.Context, event *entity.Event) error {
	var err error

	switch event.GetType() {
	case entity.OrderBeingFilled:
		order := event.GetOrder()
		orderKey := fmt.Sprintf("order:%s", order.Id)
		bookKey := fmt.Sprintf("book:%s", string(order.Type))

		pipe := this.conn.Pipeline()
		pipe.HSet(ctx, orderKey, "filled_quantity", event.NewFilledQty)
		pipe.HIncrBy(ctx, orderKey, "reserve", -event.ReserveDelta)
		pipe.EvalSha(
			ctx,
			this.incrLevelSha,
			[]string{bookKey},
			strconv.Itoa(int(order.Price)),
			-event.FilledDelta,
		)

		_, err = pipe.Exec(ctx)

	case entity.OrderStatusChanged:
		orderKey := fmt.Sprintf("order:%s", event.OrderId)
		err = this.conn.HSet(ctx, orderKey, "status", string(event.Status)).Err()

	case entity.OrderCancelled, entity.OrderRejected, entity.OrderFilled:
		order := event.GetOrder()
		orderKey := fmt.Sprintf("order:%s", order.Id)
		bookKey := fmt.Sprintf("book:%s", string(order.Type))
		orderBookMember := fmt.Sprintf("%d:%s", order.CreatedAt.UnixNano(), order.Id)

		pipe := this.conn.Pipeline()
		pipe.Del(ctx, orderKey)
		pipe.ZRem(ctx, string(order.Type), orderBookMember)
		pipe.EvalSha(
			ctx,
			this.incrLevelSha,
			[]string{bookKey},
			strconv.Itoa(int(order.Price)),
			-event.Delta,
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
	order.TimeInForce = entity.TIF(fields["time_in_force"])
	order.Price, _ = strconv.ParseInt(fields["price"], 10, 64)
	stopStatus, _ := strconv.ParseBool(fields["stop"])
	order.SetStopStatus(stopStatus)
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
