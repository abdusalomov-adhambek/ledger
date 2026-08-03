# Ledger Service

A simple double-entry ledger service written in Go using **gRPC** and **grpc-gateway**.

---

## Features

- Create account
- Transfer money between accounts
- Double-entry accounting
- Transaction history
- Reverse transaction
- Idempotency support
- Outbox pattern
- Kafka integration
- PostgreSQL persistence
- gRPC API
- REST API via grpc-gateway
- Docker support

---

## Architecture

```text
               REST Client
                    │
                    ▼
            gRPC Gateway
                    │
                    ▼
              gRPC Service
                    │
                    ▼
           Application Layer
                    │
                    ▼
              Domain Layer
                    │
                    ▼
           Repository Layer
                    │
                    ▼
               PostgreSQL

                    │
                    ▼
             Outbox Worker
                    │
                    ▼
                  Kafka
```

---

## Tech Stack

- Go
- gRPC
- grpc-gateway
- Protocol Buffers (proto3)
- Buf
- PostgreSQL
- pgx
- Goose
- Kafka
- Docker
- Docker Compose

---

## Database Schema

Tables:

- accounts
- transactions
- entries
- idempotency_keys
- outbox_events

---

## Project Structure

```text
api/
└── proto/

cmd/
└── server/

internal/
├── adapter/
│   ├── grpc/
│   ├── kafka/
│   └── postgres/
├── application/
├── db/
├── domain/
├── logger/
├── server/
└── worker/

migrations/

buf.yaml
buf.gen.yaml
Dockerfile
docker-compose.yml
Makefile
```

---

## Getting Started

Clone the repository

```bash
git clone <repository-url>
cd ledger
```

Start services

```bash
docker compose up --build
```

Run database migrations

```bash
make migrate-up
```

Generate protobuf files

```bash
make proto
```

Run application

```bash
make run
```

---

## REST API (grpc-gateway)

| Method | Endpoint | Description |
|---------|----------|-------------|
| POST | `/v1/account` | Create account |
| GET | `/v1/account/{id}` | Get account |
| POST | `/v1/transfer` | Transfer money |
| GET | `/v1/entry/{account_id}/history` | Get transaction history |
| POST | `/v1/transaction/{transaction_id}/reverse` | Reverse transaction |

---

## gRPC Service

```proto
service LedgerService {
    rpc CreateAccount(CreateAccountRequest) returns (CreateAccountResponse);
    rpc Transfer(TransferRequest) returns (TransferResponse);
    rpc GetAccount(GetAccountRequest) returns (GetAccountResponse);
    rpc GetEntryHistory(GetEntryHistoryRequest) returns (GetEntryHistoryResponse);
    rpc ReverseTransfer(ReverseTransferRequest) returns (ReverseTransferResponse);
}
```

---

## Example Transfer Request

```http
POST /v1/transfer
Content-Type: application/json
Idempotency-Key: 8b22cf09-4d78-48cf-a3d2-6db96d5d15a8
```

```json
{
    "from_account_id": "ACCOUNT_ID",
    "to_account_id": "ACCOUNT_ID",
    "amount": 1000,
    "description": "Payment"
}
```

---

## Design Principles

- Clean Architecture
- Repository Pattern
- Double-entry Accounting
- ACID Transactions
- Idempotency
- Outbox Pattern
- Event-driven Architecture

---

## Make Commands

```bash
make run
make proto
make migrate-up
make migrate-down
make lint
make test
```

---

## Future Improvements

- Authentication & Authorization
- Prometheus metrics
- Distributed tracing
- Unit tests
- Integration tests
- Multi-currency support
- Event consumers

---

## License

MIT
