package repository

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/mongodb/types"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
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
		return errs.NewRepositoryError("mongodb query execution error", err)
	}

	return nil
}

func (this *UserRepository) ReleaseBalance(ctx context.Context, userId string, toReserve int64) error {
	filter := bson.M{"_id": userId}

	query := bson.M{"$inc": bson.M{"reserved": -toReserve}}

	if _, err := this.collection.UpdateOne(ctx, filter, query); err != nil {
		return errs.NewRepositoryError("mongodb query execution error", err)
	}

	return nil
}

func (this *UserRepository) CreateUser(ctx context.Context, user *entity.User) error {

	if _, err := this.collection.InsertOne(ctx, types.FromDomainEntity(user)); err != nil {
		return errs.NewRepositoryError("mongodb query execution error", err)
	}
	return nil
}

func (this *UserRepository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	filter := bson.M{"login": login}

	res := this.collection.FindOne(ctx, filter)
	if res.Err() != nil {
		return nil, errs.NewRepositoryError("mongodb query execution error", res.Err())
	}

	var user types.User
	if err := res.Decode(&user); err != nil {
		return nil, errs.NewRepositoryError("mongodb query execution error", err)
	}

	return user.ToDomainEntity(), nil

}
