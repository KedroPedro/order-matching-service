package main

import (
	"net"
	"os"

	"github.com/KedroPedro/order-matching-engine/pkg/errs"
	"github.com/KedroPedro/order-matching-engine/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

const (
	grpcAddr = "GRPC_ADDR"
)

func SetupServer(handler proto.EngineServiceServer) {
	gSrv := grpc.NewServer()

	proto.RegisterEngineServiceServer(gSrv, handler)

	addr := os.Getenv(grpcAddr)
	if addr == "" {
		panic(errs.NewMissedEnvironmentVariableError(grpcAddr))
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	if err := gSrv.Serve(listener); err != nil {
		log.Err(err).Send()
	}
}
