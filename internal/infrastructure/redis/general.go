package redis

import (
	"context"
	"errors"
	"os"
	"strconv"

	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/redis/repository"
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
		return nil, errors.New("environment variable missed") //TODO: add normal error
	}

	var pass string
	if pass = os.Getenv(redisPassword); pass == "" {
		return nil, errors.New("environment variable missed") //TODO: add normal error
	}

	var db int
	var err error
	if dbStr := os.Getenv(redisDB); dbStr == "" {
		return nil, errors.New("environment variable missed") //TODO: add normal error
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

	if err = client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (this *Client) NewMarketRepository() interfaces.MarketStateRepository {
	return repository.NewOrderRepository(this.client.Conn())
}
