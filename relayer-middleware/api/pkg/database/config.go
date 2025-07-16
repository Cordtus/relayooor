package database

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"relayooor/api/pkg/logging"
)

type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

func NewConnection(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true, // Prepare statements for better performance
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool with auto-scaling
	baseConns := cfg.MaxOpenConns
	sqlDB.SetMaxOpenConns(baseConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Start connection pool monitor
	go monitorAndScalePool(sqlDB, baseConns)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func monitorAndScalePool(db *sql.DB, baseConns int) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	consecutiveHighLoad := 0
	consecutiveLowLoad := 0

	for range ticker.C {
		stats := db.Stats()
		utilization := float64(stats.InUse) / float64(stats.MaxOpenConnections)

		if utilization > 0.8 && stats.WaitCount > 0 {
			consecutiveHighLoad++
			consecutiveLowLoad = 0

			if consecutiveHighLoad >= 3 {
				newMax := int(float64(stats.MaxOpenConnections) * 1.5)
				if newMax > baseConns*3 {
					newMax = baseConns * 3 // Cap at 3x base
				}

				db.SetMaxOpenConns(newMax)
				logging.Info("Scaled up connection pool",
					zap.Int("new_max", newMax),
					zap.Float64("utilization", utilization))
				consecutiveHighLoad = 0
			}
		} else if utilization < 0.3 {
			consecutiveLowLoad++
			consecutiveHighLoad = 0

			if consecutiveLowLoad >= 5 {
				newMax := int(float64(stats.MaxOpenConnections) * 0.7)
				if newMax < baseConns {
					newMax = baseConns // Don't go below base
				}

				db.SetMaxOpenConns(newMax)
				logging.Info("Scaled down connection pool",
					zap.Int("new_max", newMax),
					zap.Float64("utilization", utilization))
				consecutiveLowLoad = 0
			}
		} else {
			consecutiveHighLoad = 0
			consecutiveLowLoad = 0
		}
	}
}

func DefaultConfig() Config {
	return Config{
		Host:            getEnvOrDefault("DB_HOST", "localhost"),
		Port:            getEnvAsIntOrDefault("DB_PORT", 5432),
		User:            getEnvOrDefault("DB_USER", "postgres"),
		Password:        getEnvOrDefault("DB_PASSWORD", ""),
		Database:        getEnvOrDefault("DB_NAME", "relayooor"),
		SSLMode:         getEnvOrDefault("DB_SSLMODE", "disable"),
		MaxOpenConns:    getEnvAsIntOrDefault("DB_MAX_OPEN_CONNS", 25),
		MaxIdleConns:    getEnvAsIntOrDefault("DB_MAX_IDLE_CONNS", 5),
		ConnMaxLifetime: getEnvAsDurationOrDefault("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		ConnMaxIdleTime: getEnvAsDurationOrDefault("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}