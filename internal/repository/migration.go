package repository

import (
	"context"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/sirupsen/logrus"
)

func RunMigrations(ctx context.Context, dsn string, log *logrus.Logger) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error before migrations: %w", err)
	}

	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		log.WithError(err).Error("Failed to initialize migrator")
		return fmt.Errorf("migrate.New error: %w", err)
	}
	defer func() {
		if _, err := m.Close(); err != nil {
			log.WithError(err).Warn("Failed to close migration connection")
		}
	}()

	log.Info("Applying database migrations...")

	err = m.Up()
	switch {
	case err == nil:
		log.Info("All migrations applied successfully")
		return nil
	case err == migrate.ErrNoChange:
		log.Info("No new migrations to apply")
		return nil
	default:
		log.WithError(err).Error("Migration failed")
		return fmt.Errorf("migration up error: %w", err)
	}
}
