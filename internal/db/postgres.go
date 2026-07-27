package db

import (
	"context"
	"fmt"
	"ledger/internal/config"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.PostgresConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	ctx := context.Background()

	postgresUrl := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB)

	pool, err := pgxpool.New(ctx, postgresUrl)
	if err != nil {
		logger.Error("failed to connect to postgres", err)
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping postgres", err)
		return nil, err
	}

	return pool, nil
}
