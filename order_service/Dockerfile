FROM golang:alpine AS builder
WORKDIR /build

RUN apk add --no-cache make

COPY go.mod go.sum Makefile ./
COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY crt/ ./crt/
COPY order_service/ ./order_service/

RUN make build-order

FROM alpine:latest
WORKDIR /app
COPY --from=builder /build/bin/order ./order
COPY --from=builder /build/crt/ ./crt/
EXPOSE 8080
CMD ["./order"]
