# Ledger Service

## Features

- Open Account
- Transfer
- Double Entry Accounting
- Transaction History
- Reverse Transaction
- Idempotency

## Architecture

HTTP
↓

Application
↓

Domain
↓

Repository
↓

PostgreSQL

## Tech Stack

Go
Gin
PostgreSQL
pgx
Docker

## Database Schema

(accounts, transactions, entries, idempotency_keys)

## Running

docker compose up --build

make run

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/accounts` | Create account |
| GET | `/accounts/:id` | Get account by ID |
| POST | `/transfer` | Transfer money |
| GET | `/accounts/:id/history` | Get transaction history |
| POST | `/transactions/:id/reverse` | Reverse transaction |
