package engine

import "github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/engine/repository"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (this *Client) NewEngine() (*repository.OrdersRepository, error) {
	repo, err := repository.NewOrdersRepository()
	if err != nil {
		return nil, err
	}

	return repo, nil
}
