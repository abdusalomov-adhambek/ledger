DB_URL=postgres://postgres:postgres@localhost:5433/ledger?sslmode=disable
.PHONY: run build test lint proto format clean

run:
	go run ./cmd/server

kafka:
	go run ./cmd/consumer

build:
	go build ./cmd/server

test:
	go test ./...

lint:
	golangci-lint run

down:
	docker compose down

restart:
	docker compose down
	docker compose up --build -d

proto:
	buf generate

logs:
	docker compose logs -f

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status
