FROM golang:alpine AS builder
WORKDIR /build

RUN apk add --no-cache make

COPY go.mod go.sum Makefile ./
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY engine_service/ ./engine_service/

RUN make build-engine

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/bin/engine ./engine
EXPOSE 50051
CMD ["./engine"]
