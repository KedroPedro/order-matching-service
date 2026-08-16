package handlers

import (
	"context"

	"github.com/KedroPedro/order-matching-engine/engine_service/internal/application/usecases"
	"github.com/KedroPedro/order-matching-engine/engine_service/internal/controller/converters"
	"github.com/KedroPedro/order-matching-engine/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrdersHandler struct {
	proto.UnimplementedEngineServiceServer
	addOrderUsecase    *usecases.AddOrderUsecase
	cancelOrderUsecase *usecases.CancelOrderUsecase
}

func NewOrdersHandler(
	addUsecase *usecases.AddOrderUsecase,
	cancelUsecase *usecases.CancelOrderUsecase,
) proto.EngineServiceServer {
	return &OrdersHandler{
		addOrderUsecase:    addUsecase,
		cancelOrderUsecase: cancelUsecase,
	}
}

func (this *OrdersHandler) AddOrder(ctx context.Context, req *proto.AddRequest) (*emptypb.Empty, error) {
	order, err := converters.OrderFromProtoToDomain(req.Order)
	if err != nil {
		log.Err(err).Send()
		return &emptypb.Empty{}, status.Error(codes.InvalidArgument, "convert proto order to domain error")
	}

	this.addOrderUsecase.Execute(order)

	return &emptypb.Empty{}, nil
}

func (this *OrdersHandler) CancelOrder(ctx context.Context, req *proto.CancelRequest) (*emptypb.Empty, error) {
	this.cancelOrderUsecase.Execute(req.OrderId)

	return &emptypb.Empty{}, nil
}
