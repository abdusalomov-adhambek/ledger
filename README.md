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

POST /accounts, 
GET /accounts/:id, 
POST /transfer, 
GET /accounts/:id/history, 
POST /transactions/:id/reverse
