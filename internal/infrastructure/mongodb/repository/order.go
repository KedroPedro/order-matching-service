package repository

import (
	"context"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/mongodb/mongotypes"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
	modelsCh   chan mongo.WriteModel
	sendTicker *time.Ticker
}

const (
	modelsChanBuffer     = 1000
	sendTickerIntervalMS = 300
	maxSendRetryValue    = 5
)

func New(ctx context.Context, orderCollection *mongo.Collection) *OrderRepository {
	newOrderRepository := &OrderRepository{
		collection: orderCollection,
		modelsCh:   make(chan mongo.WriteModel, modelsChanBuffer),
		sendTicker: time.NewTicker(time.Millisecond * sendTickerIntervalMS),
	}

	go func(ctx context.Context) {
		models := make([]mongo.WriteModel, 0, modelsChanBuffer)

		for {
			select {
			case model, ok := <-newOrderRepository.modelsCh:
				if !ok {
					newOrderRepository.sendBulk(ctx, models)
					return
				}
				models = append(models, model)

			case <-newOrderRepository.sendTicker.C:
				if len(models) == 0 {
					continue
				}

				bulk := make([]mongo.WriteModel, len(models))
				copy(bulk, models)

				go func(bulk []mongo.WriteModel) {
					if err := newOrderRepository.sendBulk(ctx, bulk); err != nil {
						//TODO: add logger
					}
				}(bulk)

				models = models[:0]

			case <-ctx.Done():
				newOrderRepository.sendBulk(ctx, models)
				return
			}
		}
	}(ctx)

	return newOrderRepository
}

func (this *OrderRepository) sendBulk(ctx context.Context, bulk []mongo.WriteModel) error {
	var err error = nil
	for i := 0; i <= maxSendRetryValue; i++ {
		time.Sleep(time.Millisecond * 100 * time.Duration(i*i))

		if _, err = this.collection.BulkWrite(ctx, bulk); err != nil {
			//TODO: add logging
			continue
		} else {
			break
		}
	}

	return err
}

func (this *OrderRepository) AddToQueue(ctx context.Context, order *entity.Order) error {
	insert := createOrderInsertModel(order)

	select {
	case this.modelsCh <- insert:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (this *OrderRepository) ProcessEvent(ctx context.Context, event *entity.Event) error {
	models := make([]mongo.WriteModel, 0, 2)

	models = append(models, createEventInsertModel(event))

	models = append(models, createEventUpdateModel(event))

	for _, model := range models {
		select {
		case this.modelsCh <- model:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func createEventInsertModel(event *entity.Event) mongo.WriteModel {
	return mongo.NewInsertOneModel().SetDocument(bson.D{
		bson.E{Key: "_id", Value: uuid.New().String()},
		bson.E{Key: "order_id", Value: event.GetOrderId()},
		bson.E{Key: "value", Value: event.GetValue()},
		bson.E{Key: "type", Value: event.GetType()},
	})
}

func createEventUpdateModel(event *entity.Event) mongo.WriteModel { // TODO: need some changes
	filter := bson.M{"_id": event.GetOrderId()}
	var update bson.M

	switch event.GetType() {
	case entity.OrderStatusChanged:
		update = bson.M{"$set": bson.M{"status": event.GetValue()}}

	case entity.OrderReserveChanged:
		update = bson.M{"$inc": bson.M{"reserve": -event.GetValue().(int)}}

	case entity.OrderBeingFilled:
		update = bson.M{"$inc": bson.M{"filled_quantity": event.GetValue().(int)}}
	}

	return mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
}

func createOrderInsertModel(order *entity.Order) mongo.WriteModel {
	mongoOrder := mongotypes.OrderFromDomain(order)

	return mongo.NewInsertOneModel().SetDocument(mongoOrder)
}
