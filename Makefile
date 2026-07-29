DB_URL=postgres://postgres:postgres@localhost:5433/ledger?sslmode=disable

run:
	go run ./cmd/server

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

logs:
	docker compose logs -f

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir migrations postgres "$(DB_URL)" status
