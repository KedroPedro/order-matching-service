
build-engine:
	GCO_ENABLED=0 go build -o bin/engine ./engine_service/cmd/engine_service/*

build-order:
	GCO_ENABLED=0 go build -o bin/order ./order_service/cmd/app/*
