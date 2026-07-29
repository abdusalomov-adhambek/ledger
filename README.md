# Ledger Service

A simple double-entry ledger service written in Go.

## Features

- Open account
- Transfer money between accounts
- Double-entry accounting
- Transaction history
- Reverse transaction
- Idempotency support
- Outbox pattern
- PostgreSQL persistence
- Docker support

---

## Architecture

```text
                HTTP API (Gin)
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
```

---

## Tech Stack

- Go
- Gin
- PostgreSQL
- pgx
- Goose
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
cmd/
internal/
    adapter/
    application/
    domain/
    handler/
    repository/
    db/
migrations/
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

Start PostgreSQL and API

```bash
make up
```

Run database migrations

```bash
make migrate-up
```

Stop containers

```bash
make down
```

View logs

```bash
make logs
```

---

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/accounts` | Create account |
| GET | `/accounts/:id` | Get account |
| POST | `/transfer` | Transfer money |
| GET | `/accounts/:id/history` | Get account transaction history |
| POST | `/transactions/:id/reverse` | Reverse transaction |

---

## Example Transfer Request

```http
POST /transfer
```

```json
{
    "from_account_id": "ACCOUNT_ID",
    "to_account_id": "ACCOUNT_ID",
    "amount": 1000,
    "currency": "USD"
}
```

---

## Design Principles

- Double-entry accounting
- ACID transactions
- Idempotent requests
- Optimistic architecture
- Clean Architecture
- Repository Pattern
- Outbox Pattern

---

## Make Commands

```bash
make up
make down
make restart
make logs
make migrate-up
make migrate-down
make migrate-status
```

---

## Future Improvements

- Background Outbox Worker
- Kafka integration
- gRPC API
- Authentication & Authorization
- Metrics (Prometheus)
- Distributed tracing
- Unit and Integration tests

---

## License

MIT


                 Client
                    │
                    ▼
              Gin HTTP Server
                    │
                    ▼
            Application Layer
                    │
        ┌───────────┴───────────┐
        ▼                       ▼
 Account Repository      Outbox Repository
        │                       │
        └───────────┬───────────┘
                    ▼
               PostgreSQL
