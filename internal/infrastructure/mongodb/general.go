package mongodb

import (
	"context"
	"errors"
	"os"

	"github.com/KedroPedro/order-matching-engine/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/internal/infrastructure/mongodb/repository"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

const (
	mongodbUri              = "MONGODB_URI"
	mongodbOrdersCollection = "orders"
	mongodbUsersCollection  = "users"
	mongodbDatabase         = "OrderMatchingService"
)

func NewMongoClient() (*Client, error) {
	var uri string
	if uri = os.Getenv(mongodbUri); uri == "" {
		return nil, errors.New("environment variable missed") //TODO: add normal error
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
		db:     client.Database(mongodbDatabase),
	}, nil
}

func (this *Client) NewOrderRepository(ctx context.Context) interfaces.OrderRepository {
	return repository.NewOrderRepository(ctx, this.db.Collection(mongodbOrdersCollection))
}

func (this *Client) NewUserRepository() interfaces.UserRepository {
	return repository.NewUserRepository(this.db.Collection(mongodbUsersCollection))
}
