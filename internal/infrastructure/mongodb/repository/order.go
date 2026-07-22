package repository

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
	modelsCh   chan mongo.WriteModel
}

func New(orderCollection *mongo.Collection) *OrderRepository {

	return &OrderRepository{
		collection: orderCollection,
	}
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
	mongo.NewInsertOneModel()
	return nil
}

func createEventUpdateModel(event *entity.Event) mongo.WriteModel {
	mongo.NewUpdateOneModel()
	return nil
}

func createOrderInsertModel(order *entity.Order) mongo.WriteModel {
	return mongo.NewInsertOneModel().SetDocument(order)
}
