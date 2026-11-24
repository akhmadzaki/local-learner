package db

import (
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // registers the "postgres" driver
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"github.com/akhmadzaki/go-backend/config"
)

// migrationsFS embeds every *.sql migration into the compiled binary so the
// running server (or migrate CLI) needs no on-disk migration files.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// DSN returns a postgres:// connection URL compatible with golang-migrate.
func DSN(cfg *config.DatabaseConfig) string {
	return fmt.Sprintf(
		"postgres://%s@%s:%d/%s?sslmode=disable",
		url.UserPassword(cfg.User, cfg.Password),
		cfg.Host, cfg.Port, cfg.Name,
	)
}

// newMigrator builds a golang-migrate.Migrate instance backed by the
// embedded SQL migrations.
func newMigrator(cfg *config.DatabaseConfig) (*migrate.Migrate, error) {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("db: load embedded migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, DSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("db: init migrator: %w", err)
	}
	return m, nil
}

// MigrateUp applies all pending migrations. It returns nil when there is
// nothing left to apply (migrate.ErrNoChange).
func MigrateUp(cfg *config.DatabaseConfig) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: migrate up: %w", err)
	}
	return nil
}

// MigrateDown rolls back the most recently applied migration (one step).
func MigrateDown(cfg *config.DatabaseConfig) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: migrate down: %w", err)
	}
	return nil
}

// MigrateTo migrates (up or down) to the given schema version.
func MigrateTo(cfg *config.DatabaseConfig, version uint) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Migrate(version); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: migrate to %d: %w", version, err)
	}
	return nil
}

// MigrateDrop runs every down migration, dropping all managed tables.
// Use with care — it destroys data.
func MigrateDrop(cfg *config.DatabaseConfig) error {
	m, err := newMigrator(cfg)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Drop(); err != nil {
		return fmt.Errorf("db: migrate drop: %w", err)
	}
	return nil
}

// MigrateVersion reports the current schema version and whether the
// migration state is dirty (an interrupted migration).
func MigrateVersion(cfg *config.DatabaseConfig) (version uint, dirty bool, err error) {
	m, err := newMigrator(cfg)
	if err != nil {
		return 0, false, err
	}
	defer m.Close()

	version, dirty, err = m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("db: migrate version: %w", err)
	}
	return version, dirty, nil
}
