package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// PoolConfig contains database pool configuration
type PoolConfig struct {
	// Connection settings
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string

	// Pool settings
	MaxConnections        int32         // Maximum number of connections in the pool
	MinConnections        int32         // Minimum number of connections to maintain
	MaxConnLifetime       time.Duration // Maximum lifetime of a connection
	MaxConnIdleTime       time.Duration // Maximum idle time before closing a connection
	HealthCheckPeriod     time.Duration // How often to check connection health
	ConnectionTimeout     time.Duration // Timeout for acquiring a connection
	StatementCacheEnabled bool          // Enable prepared statement caching
	
	// Performance settings
	DefaultQueryTimeout   time.Duration // Default timeout for queries
	SlowQueryThreshold    time.Duration // Threshold for logging slow queries
}

// DefaultPoolConfig returns sensible defaults for the connection pool
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		Host:                  "localhost",
		Port:                  5432,
		Database:              "relayooor",
		Username:              "postgres",
		Password:              "postgres",
		SSLMode:               "disable",
		MaxConnections:        25,
		MinConnections:        5,
		MaxConnLifetime:       time.Hour,
		MaxConnIdleTime:       time.Minute * 30,
		HealthCheckPeriod:     time.Minute,
		ConnectionTimeout:     time.Second * 30,
		StatementCacheEnabled: true,
		DefaultQueryTimeout:   time.Second * 30,
		SlowQueryThreshold:    time.Second * 5,
	}
}

// Pool wraps pgxpool with additional functionality
type Pool struct {
	*pgxpool.Pool
	config *PoolConfig
	logger *zap.Logger
	sqlDB  *sql.DB
}

// NewPool creates a new database connection pool
func NewPool(ctx context.Context, config *PoolConfig, logger *zap.Logger) (*Pool, error) {
	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database, config.SSLMode,
	)

	// Parse config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply pool settings
	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectionTimeout

	// Set up query logging for slow queries
	poolConfig.BeforeAcquire = func(ctx context.Context, conn *pgxpool.Conn) bool {
		return true // Could add custom logic here
	}

	poolConfig.AfterRelease = func(conn *pgxpool.Conn) bool {
		return true // Could add custom logic here
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create sql.DB interface for compatibility
	sqlDB := stdlib.OpenDBFromPool(pool)

	logger.Info("Database pool created",
		zap.Int32("max_connections", config.MaxConnections),
		zap.Int32("min_connections", config.MinConnections),
		zap.String("database", config.Database),
	)

	return &Pool{
		Pool:   pool,
		config: config,
		logger: logger,
		sqlDB:  sqlDB,
	}, nil
}

// DB returns a sql.DB interface for compatibility with existing code
func (p *Pool) DB() *sql.DB {
	return p.sqlDB
}

// ExecuteWithTimeout executes a query with a timeout
func (p *Pool) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	result, err := p.Exec(ctx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > p.config.SlowQueryThreshold {
		p.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Any("args", args),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	return &pgxResult{result}, nil
}

// QueryWithTimeout executes a query with a timeout and returns rows
func (p *Pool) QueryWithTimeout(ctx context.Context, timeout time.Duration, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	rows, err := p.Query(ctx, query, args...)
	duration := time.Since(start)

	// Log slow queries
	if duration > p.config.SlowQueryThreshold {
		p.logger.Warn("Slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
			zap.Any("args", args),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// Convert pgx rows to sql.Rows
	return stdlib.RowsFromRows(rows), nil
}

// Transaction executes a function within a database transaction
func (p *Pool) Transaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := p.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction failed and rollback failed: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HealthCheck performs a health check on the pool
func (p *Pool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var result int
	err := p.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	stats := p.Stat()
	p.logger.Debug("Pool health check",
		zap.Int32("acquired_conns", stats.AcquiredConns()),
		zap.Int32("idle_conns", stats.IdleConns()),
		zap.Int64("total_conns", stats.TotalConns()),
		zap.Int64("new_conns_count", stats.NewConnsCount()),
	)

	return nil
}

// WaitForNotification waits for a PostgreSQL notification
func (p *Pool) WaitForNotification(ctx context.Context, channel string, handler func(payload string)) error {
	conn, err := p.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Listen to the channel
	_, err = conn.Exec(ctx, "LISTEN "+channel)
	if err != nil {
		return fmt.Errorf("failed to listen to channel %s: %w", channel, err)
	}

	// Wait for notifications
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return fmt.Errorf("error waiting for notification: %w", err)
			}
			handler(notification.Payload)
		}
	}
}

// Close closes the connection pool
func (p *Pool) Close() {
	p.Pool.Close()
	if p.sqlDB != nil {
		_ = p.sqlDB.Close()
	}
	p.logger.Info("Database pool closed")
}

// pgxResult wraps pgx result to implement sql.Result
type pgxResult struct {
	pgx.CommandTag
}

func (r *pgxResult) LastInsertId() (int64, error) {
	return 0, fmt.Errorf("LastInsertId not supported by PostgreSQL")
}

func (r *pgxResult) RowsAffected() (int64, error) {
	return r.CommandTag.RowsAffected(), nil
}