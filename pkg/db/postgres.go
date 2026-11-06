package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

// Connect initializes a PostgreSQL connection pool with timeout
func Connect(databaseURL string) (*DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("connect db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	log.Println("[Database] Successfully connected to PostgreSQL 16 pool")
	return &DB{Pool: pool}, nil
}

// Close closes the pool
func (db *DB) Close() {
	if db != nil && db.Pool != nil {
		db.Pool.Close()
	}
}

// Migrate executes the SQL migration file if exists
func (db *DB) Migrate(schemaPath string) error {
	if db == nil || db.Pool == nil {
		return nil
	}

	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = db.Pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	log.Println("[Database] Migration applied successfully:", schemaPath)
	return nil
}

// migrateLockKey serializes concurrent migrators (parallel tests, replicas)
// around DDL such as CREATE EXTENSION, whose IF NOT EXISTS still races.
const migrateLockKey = 72771617

// MigrateAll executes every *.sql file in dir in sorted filename order.
func (db *DB) MigrateAll(dir string) error {
	if db == nil || db.Pool == nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migrate lock holder: %w", err)
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, int64(migrateLockKey)); err != nil {
		return fmt.Errorf("acquire migrate lock: %w", err)
	}
	defer func() { _, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, int64(migrateLockKey)) }()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if err := db.Migrate(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}
