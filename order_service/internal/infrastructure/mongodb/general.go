package mongodb

import (
	"context"
	"os"
	"time"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/interfaces"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/mongodb/repository"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

const (
	mongodbUri              = "MONGODB_URI"
	mongodbOrdersCollection = "orders"
	mongodbEventsCollection = "events"
	mongodbUsersCollection  = "users"
	mongodbDatabase         = "matching_db"
)

func NewMongoClient() (*Client, error) {
	var uri string
	if uri = os.Getenv(mongodbUri); uri == "" {
		return nil, errs.NewMissedEnvironmentVariableError(mongodbUri)
	}

	connCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(connCtx, readpref.Primary()); err != nil {
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
