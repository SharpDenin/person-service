package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"person-service/internal/config"
)

func NewDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logrus.Errorf("Failed to connect to database: %v", err)
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		logrus.Errorf("Failed to ping database: %v", err)
		pool.Close()
		return nil, err
	}

	logrus.Info("Successfully connected to database")
	return pool, nil
}
