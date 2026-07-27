run:
	go run ./cmd/server

build:
	go build ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run

migrate-up:

migrate-down:
