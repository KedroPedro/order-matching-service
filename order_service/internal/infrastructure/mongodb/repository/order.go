package repository

import (
	"context"
	"time"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/mongodb/types"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	orderCollection *mongo.Collection
	orderModelsCh   chan mongo.WriteModel
	sendTicker      *time.Ticker
}

const (
	modelsChanBuffer     = 100000
	sendTickerIntervalMS = 300
	maxSendRetryValue    = 5
)

func NewOrderRepository(ctx context.Context, orderCollection *mongo.Collection) *OrderRepository {
	newOrderRepository := &OrderRepository{
		orderCollection: orderCollection,
		orderModelsCh:   make(chan mongo.WriteModel, modelsChanBuffer),
		sendTicker:      time.NewTicker(time.Millisecond * sendTickerIntervalMS),
	}

	go func(ctx context.Context) {
		orderModels := make([]mongo.WriteModel, 0, modelsChanBuffer)

		for {
			select {
			case model, ok := <-newOrderRepository.orderModelsCh:
				if !ok {
					sendBulk(ctx, newOrderRepository.orderCollection, orderModels)
					return
				}
				orderModels = append(orderModels, model)

			case <-newOrderRepository.sendTicker.C:

				if len(orderModels) != 0 {
					ordersBulk := make([]mongo.WriteModel, len(orderModels))
					copy(ordersBulk, orderModels)
					if err := sendBulk(ctx, newOrderRepository.orderCollection, ordersBulk); err != nil {
						log.Err(err).Send()
					}

					orderModels = orderModels[:0]
				}

			case <-ctx.Done():
				sendBulk(ctx, newOrderRepository.orderCollection, orderModels)
				return
			}
		}
	}(ctx)

	return newOrderRepository
}

func sendBulk(ctx context.Context, collection *mongo.Collection, bulk []mongo.WriteModel) error {
	var err error = nil
	for i := 0; i <= maxSendRetryValue; i++ {
		time.Sleep(time.Millisecond * 100 * time.Duration(i*i))

		if _, err = collection.BulkWrite(ctx, bulk); err != nil {
			log.Err(err).Send()
			continue
		} else {
			break
		}
	}

	return err
}

func (this *OrderRepository) AddOrder(ctx context.Context, order *entity.Order) error {
	insert := createOrderInsertModel(order)

	select {
	case this.orderModelsCh <- insert:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func createOrderInsertModel(order *entity.Order) mongo.WriteModel {
	mongoOrder := types.OrderFromDomain(order)

	return mongo.NewInsertOneModel().SetDocument(mongoOrder)
}
