FROM golang:alpine AS builder
WORKDIR /build

RUN apk add --no-cache make

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app
COPY --from=builder /build/app /app/app
COPY --from=builder /build/crt/ /app/crt/

CMD ["./app"]
