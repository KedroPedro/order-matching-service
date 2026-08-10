package repository

import (
	"context"
	"sync"
	"time"

	"github.com/KedroPedro/order-matching-engine/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/mongodb/types"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	orderCollection *mongo.Collection
	eventCollection *mongo.Collection
	orderModelsCh   chan mongo.WriteModel
	eventModelsCh   chan mongo.WriteModel
	sendTicker      *time.Ticker
}

const (
	modelsChanBuffer     = 100000
	sendTickerIntervalMS = 300
	maxSendRetryValue    = 5
)

func NewOrderRepository(ctx context.Context, orderCollection *mongo.Collection, eventCollection *mongo.Collection) *OrderRepository {
	newOrderRepository := &OrderRepository{
		orderCollection: orderCollection,
		eventCollection: eventCollection,
		orderModelsCh:   make(chan mongo.WriteModel, modelsChanBuffer),
		eventModelsCh:   make(chan mongo.WriteModel, modelsChanBuffer),
		sendTicker:      time.NewTicker(time.Millisecond * sendTickerIntervalMS),
	}

	go func(ctx context.Context) {
		orderModels := make([]mongo.WriteModel, 0, modelsChanBuffer)
		eventModels := make([]mongo.WriteModel, 0, modelsChanBuffer)
		wg := sync.WaitGroup{}

		for {
			select {
			case model, ok := <-newOrderRepository.orderModelsCh:
				if !ok {
					wg.Go(func() {
						sendBulk(ctx, newOrderRepository.orderCollection, orderModels)
					})
					wg.Go(func() {
						sendBulk(ctx, newOrderRepository.eventCollection, eventModels)
					})
					wg.Wait()
					return
				}
				orderModels = append(orderModels, model)

			case model, ok := <-newOrderRepository.eventModelsCh:
				if !ok {
					wg.Go(func() {
						sendBulk(ctx, newOrderRepository.orderCollection, orderModels)
					})
					wg.Go(func() {
						sendBulk(ctx, newOrderRepository.eventCollection, eventModels)
					})
					wg.Wait()
					return
				}
				eventModels = append(eventModels, model)

			case <-newOrderRepository.sendTicker.C:

				if len(orderModels) != 0 {
					ordersBulk := make([]mongo.WriteModel, len(orderModels))
					copy(ordersBulk, orderModels)
					wg.Go(func() {
						if err := sendBulk(ctx, newOrderRepository.orderCollection, ordersBulk); err != nil {
							log.Err(err).Send()
						}
					})

					orderModels = orderModels[:0]
				}

				if len(eventModels) != 0 {
					eventsBulk := make([]mongo.WriteModel, len(eventModels))
					copy(eventsBulk, eventModels)

					wg.Go(func() {
						if err := sendBulk(ctx, newOrderRepository.eventCollection, eventsBulk); err != nil {
							log.Err(err).Send()
						}
					})

					eventModels = eventModels[:0]
				}

				wg.Wait()

			case <-ctx.Done():
				wg.Go(func() {
					sendBulk(ctx, newOrderRepository.orderCollection, orderModels)
				})
				wg.Go(func() {
					sendBulk(ctx, newOrderRepository.eventCollection, eventModels)
				})
				wg.Wait()
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

func (this *OrderRepository) AddToQueue(ctx context.Context, order *entity.Order) error {
	insert := createOrderInsertModel(order)

	select {
	case this.orderModelsCh <- insert:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (this *OrderRepository) ProcessEvent(ctx context.Context, event *entity.Event) error {
	insertModel := createEventInsertModel(event)
	select {
	case this.eventModelsCh <- insertModel:
	case <-ctx.Done():
		return ctx.Err()
	}

	updateModel := createEventUpdateModel(event)
	select {
	case this.orderModelsCh <- updateModel:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
func createEventInsertModel(event *entity.Event) mongo.WriteModel {
	var orderId string
	var value bson.M

	switch event.GetType() {
	case entity.OrderBeingFilled:
		orderId = event.GetOrderId()
		value = bson.M{
			"filled_delta":   event.FilledDelta,
			"new_filled_qty": event.NewFilledQty,
			"reserve_delta":  event.ReserveDelta,
		}

	case entity.OrderQuantityChanged:
		orderId = event.OrderId
		value = bson.M{
			"quantity_delta": event.Quantity,
			"price":          event.Price,
			"order_type":     string(event.OrderType),
		}

	case entity.OrderStatusChanged:
		orderId = event.OrderId
		value = bson.M{
			"status": string(event.Status),
		}

	case entity.OrderCancelled, entity.OrderRejected, entity.OrderFilled:
		orderId = event.GetOrderId()
		value = bson.M{
			"delta": event.Delta,
		}

	case entity.OrderReserveChanged:
		orderId = event.OrderId
		value = bson.M{
			"reserve_delta": event.ReserveDelta,
		}
	}

	return mongo.NewInsertOneModel().SetDocument(bson.D{
		bson.E{Key: "_id", Value: uuid.New().String()},
		bson.E{Key: "order_id", Value: orderId},
		bson.E{Key: "value", Value: value},
		bson.E{Key: "type", Value: event.GetType()},
	})
}

func createEventUpdateModel(event *entity.Event) mongo.WriteModel {
	var orderId string
	var update bson.M

	switch event.Type {
	case entity.OrderBeingFilled:
		orderId = event.GetOrderId()
		update = bson.M{
			"$inc": bson.M{
				"filled_quantity": event.FilledDelta,
				"reserve":         -event.ReserveDelta,
			},
		}

	case entity.OrderQuantityChanged:
		orderId = event.OrderId
		update = bson.M{
			"$inc": bson.M{
				"quantity": event.Quantity,
			},
		}

	case entity.OrderStatusChanged:
		orderId = event.OrderId
		update = bson.M{
			"$set": bson.M{
				"status": string(event.Status),
			},
		}

	case entity.OrderCancelled, entity.OrderRejected, entity.OrderFilled:
		orderId = event.GetOrderId()
		update = bson.M{
			"$set": bson.M{
				"status": string(event.Status),
			},
			"$inc": bson.M{
				"filled_quantity": event.Delta,
			},
		}
	}

	filter := bson.M{"_id": orderId}
	return mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update)
}

func createOrderInsertModel(order *entity.Order) mongo.WriteModel {
	mongoOrder := types.OrderFromDomain(order)

	return mongo.NewInsertOneModel().SetDocument(mongoOrder)
}
