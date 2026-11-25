package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ryanmcafee/smart-garden-bot/api/internal/config"
)

type DB struct {
	conn *sqlx.DB
}

// Initialize creates a new database connection pool
func Initialize(cfg config.DatabaseConfig) (*DB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	// Create connection
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxConns)
	db.SetMaxIdleConns(cfg.MinConns)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(time.Minute * 30)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Note: Migrations should be run separately, not in the application
	// For now, we'll skip automatic migrations

	return &DB{conn: db}, nil
}

// HealthCheck checks if the database is healthy
func HealthCheck(ctx context.Context, db *DB) error {
	return db.PingContext(ctx)
}

// GetStats returns database statistics
func GetStats(db *DB) sql.DBStats {
	return db.Stats()
}

// Transaction executes a function within a database transaction
func Transaction(ctx context.Context, db *sqlx.DB, fn func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

// Repository interface defines common database operations
type Repository interface {
	HealthCheck(ctx context.Context) error
}

// BaseRepository provides common database functionality
type BaseRepository struct {
	db *sqlx.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sqlx.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// DB returns the database instance
func (r *BaseRepository) DB() *sqlx.DB {
	return r.db
}

// HealthCheck implements the Repository interface
func (r *BaseRepository) HealthCheck(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Database service methods

// GetContext performs a single row query and scans the result into dest
func (db *DB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.conn.GetContext(ctx, dest, query, args...)
}

// SelectContext performs a multi-row query and scans the results into dest
func (db *DB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return db.conn.SelectContext(ctx, dest, query, args...)
}

// ExecContext executes a query without returning any rows
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

// NamedExecContext executes a query with named parameters
func (db *DB) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return db.conn.NamedExecContext(ctx, query, arg)
}

// PingContext verifies the database connection is alive
func (db *DB) PingContext(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// Stats returns database connection pool statistics
func (db *DB) Stats() sql.DBStats {
	return db.conn.Stats()
}

// BeginTxx starts a transaction with context and options
func (db *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	return db.conn.BeginTxx(ctx, opts)
}

// Transaction executes a function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*sqlx.Tx) error) error {
	return Transaction(ctx, db.conn, fn)
}

// QueryContext executes a query that returns rows
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying sqlx.DB connection for advanced use cases
func (db *DB) Conn() *sqlx.DB {
	return db.conn
}