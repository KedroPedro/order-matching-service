build:
	go mod download && go build -o ./app ./cmd/app/

run:
	go run ./cmd/app/main.go

download:
	go mod download

test:
	go test -count=1 ./...
