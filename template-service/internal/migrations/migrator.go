package migrations

import (
	"fmt"
	"path/filepath"

	"template-service/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

type Migrator struct {
	logger *zap.SugaredLogger
	dsn    string
	path   string
}

func New(cfg config.PostgresConfig, logger *zap.SugaredLogger) (*Migrator, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	absPath, err := filepath.Abs("./migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve migrations path: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(absPath)
	return &Migrator{
		logger: logger,
		dsn:    dsn,
		path:   sourceURL,
	}, nil
}

func (m *Migrator) RunMigrations() error {
	m.logger.Infof("Running migrations from %s", m.path)

	migrator, err := migrate.New(m.path, m.dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("No migrations to apply.")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	version, dirty, _ := migrator.Version()
	m.logger.Infof("Migrations applied. Version: %d, dirty: %v", version, dirty)
	return nil
}

func (m *Migrator) RollbackLast() error {
	m.logger.Warnf("Rolling back last migration from %s", m.path)

	migrator, err := migrate.New(m.path, m.dsn)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Steps(-1); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	version, dirty, _ := migrator.Version()
	m.logger.Infof("Rollback complete. Version: %d, dirty: %v", version, dirty)
	return nil
}
