package repository

import (
	"context"
	"os"

	"github.com/KedroPedro/order-matching-engine/order_service/internal/domain/entity"
	"github.com/KedroPedro/order-matching-engine/order_service/internal/infrastructure/engine/converters"
	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/KedroPedro/order-matching-engine/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrdersRepository struct {
	client proto.EngineServiceClient
}

const (
	grpcTarget = "GRPC_TARGET"
)

func NewOrdersRepository() (*OrdersRepository, error) {
	target := os.Getenv(grpcTarget)
	if target == "" {
		return nil, errs.NewMissedEnvironmentVariableError(grpcTarget)
	}

	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errs.NewRepositoryError("create grpc connection error", err)
	}

	client := proto.NewEngineServiceClient(conn)

	return &OrdersRepository{
		client: client,
	}, nil
}

func (this *OrdersRepository) AddOrder(ctx context.Context, order *entity.Order) error {
	pOrder, err := converters.OrderFromDomainToProto(order)
	if err != nil {
		return errs.NewRepositoryError("convert from domain to proto order error", err)
	}

	if _, err = this.client.AddOrder(ctx, &proto.AddRequest{Order: pOrder}); err != nil {
		return errs.NewRepositoryError("add order error", err)
	}

	return nil
}

func (this *OrdersRepository) CancelOrder(ctx context.Context, orderId string) error {
	if _, err := this.client.CancelOrder(ctx, &proto.CancelRequest{OrderId: orderId}); err != nil {
		return errs.NewRepositoryError("cancel order error", err)
	}

	return nil
}
