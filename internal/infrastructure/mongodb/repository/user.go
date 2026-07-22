package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(userCollection *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: userCollection,
	}
}

func (this *UserRepository) ReserveBalance(ctx context.Context, userId string, toReserve int64) error {
	filter := bson.M{"_id": userId}

	query := bson.M{"$inc": bson.M{"reserved": toReserve}}

	if _, err := this.collection.UpdateOne(ctx, filter, query); err != nil {
		return err
	}

	return nil
}

func (this *UserRepository) ReleaseBalance(ctx context.Context, userId string, toReserve int64) error {
	filter := bson.M{"_id": userId}

	query := bson.M{"$inc": bson.M{"reserved": -toReserve}}

	if _, err := this.collection.UpdateOne(ctx, filter, query); err != nil {
		return err
	}

	return nil
}

func (this *UserRepository) UpdateBalance(ctx context.Context, userId string, reserved int64, price int64) error {
	filter := bson.M{"_id": userId}

	query := bson.M{"$inc": bson.M{"reserved": -reserved, "balance": price}}

	_, err := this.collection.UpdateOne(ctx, filter, query)
	return err

}
