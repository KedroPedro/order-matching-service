package redis

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/redis/repository"
	"github.com/KedroPedro/order-matching-engine/internal/pkg/errs"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

const (
	redisAddr     = "REDIS_ADDR"
	redisPassword = "REDIS_PASSWORD"
	redisDB       = "REDIS_DB"
)

func NewClient() (*Client, error) {
	var addr string
	if addr = os.Getenv(redisAddr); addr == "" {
		return nil, errs.NewMissedEnvironmentVariableError(redisAddr)
	}

	var pass string
	if pass = os.Getenv(redisPassword); pass == "" {
		return nil, errs.NewMissedEnvironmentVariableError(redisPassword)
	}

	var db int
	var err error
	if dbStr := os.Getenv(redisDB); dbStr == "" {
		return nil, errs.NewMissedEnvironmentVariableError(redisDB)
	} else {
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return nil, err
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err = client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (this *Client) NewMarketRepository() interfaces.MarketStateRepository {
	return repository.NewOrderRepository(this.client)
}

func (this *Client) NewSessionRepository() interfaces.SessionRepository {
	return repository.NewSessionRepository(this.client)
}
