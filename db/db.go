package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/akhmadzaki/go-backend/config"
)

// Connect establishes a PostgreSQL connection using the application config.
// It verifies the connection with a ping before returning.
func Connect(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("db.Connect: failed to open connection: %w", err)
	}

	// Verify the connection is alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Connect: ping failed: %w", err)
	}

	// Connection pool configuration from env vars with sensible defaults
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	log.Printf("db.Connect: connected to PostgreSQL at %s:%d/%s (pool: %d/%d, maxLifetime: %dm)",
		cfg.Host, cfg.Port, cfg.Name, cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
	return db, nil
}
