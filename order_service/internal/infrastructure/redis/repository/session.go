package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	conn *redis.Client
}

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{
		conn: client,
	}
}

const (
	sessionLifeTime = time.Second * 60 * 10
)

func (this *SessionRepository) GetSession(ctx context.Context, login string) (*entity.User, error) {
	key := fmt.Sprintf("user_id:%s", login)

	pipe := this.conn.Pipeline()

	res := pipe.Get(ctx, key)
	pipe.Expire(ctx, key, sessionLifeTime)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, errs.NewRepositoryError("get session error", err)
	}

	if res == nil {
		return nil, errs.NewRepositoryError("session not find", nil)
	}

	ordersSrz, err := res.Result()
	if err != nil {
		return nil, errs.NewRepositoryError("get session error", err)
	}

	var fields map[string]any

	json.Unmarshal([]byte(ordersSrz), &fields)

	user := &entity.User{
		Id:               fields["id"].(string),
		Login:            fields["login"].(string),
		TotalBalance:     fields["total_balance"].(int64),
		AvailableBalance: fields["available_balance"].(int64),
		Reserved:         fields["reserved"].(int64),
	}

	return user, nil
}

func (this *SessionRepository) AddSession(ctx context.Context, user *entity.User) error {
	key := fmt.Sprintf("user_id:%s", user.Login)

	_, err := this.conn.SetEx(ctx, key, map[string]any{
		"id":                user.Id,
		"login":             user.Login,
		"total_balance":     user.TotalBalance,
		"available_balance": user.AvailableBalance,
		"reserved":          user.Reserved,
	}, sessionLifeTime).Result()
	if err != nil {
		return errs.NewRepositoryError("save session error", err)
	}

	return nil
}
