package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IndexMonitor monitors database index usage and performance
type IndexMonitor struct {
	db     *gorm.DB
	logger *zap.Logger
}

// IndexStats represents statistics for a database index
type IndexStats struct {
	SchemaName  string
	TableName   string
	IndexName   string
	IdxScan     int64
	IdxTupRead  int64
	IdxTupFetch int64
	IndexSize   string
}

// SlowQuery represents a slow query that might benefit from indexes
type SlowQuery struct {
	Query         string
	Calls         int64
	MeanExecTime  float64
	TotalExecTime float64
	StddevTime    float64
}

// TableStats represents table statistics
type TableStats struct {
	TableName    string
	TotalRows    int64
	DeadRows     int64
	LastVacuum   *time.Time
	LastAutoVacuum *time.Time
	IndexScans   int64
	SeqScans     int64
}

// NewIndexMonitor creates a new index monitor
func NewIndexMonitor(db *gorm.DB, logger *zap.Logger) *IndexMonitor {
	return &IndexMonitor{
		db:     db,
		logger: logger.With(zap.String("component", "index_monitor")),
	}
}

// CheckIndexUsage checks which indexes are being used
func (m *IndexMonitor) CheckIndexUsage(ctx context.Context) error {
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			idx_scan,
			idx_tup_read,
			idx_tup_fetch,
			pg_size_pretty(pg_relation_size(indexrelid)) as index_size
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
		ORDER BY idx_scan DESC
	`
	
	var stats []IndexStats
	if err := m.db.WithContext(ctx).Raw(query).Scan(&stats).Error; err != nil {
		m.logger.Error("Failed to fetch index stats", zap.Error(err))
		return fmt.Errorf("failed to fetch index stats: %w", err)
	}
	
	// Log statistics
	for _, stat := range stats {
		logger := m.logger.With(
			zap.String("table", stat.TableName),
			zap.String("index", stat.IndexName),
			zap.String("size", stat.IndexSize),
			zap.Int64("scans", stat.IdxScan),
		)
		
		if stat.IdxScan == 0 {
			logger.Warn("Unused index detected")
		} else {
			logger.Debug("Index usage stats",
				zap.Int64("tuples_read", stat.IdxTupRead),
				zap.Int64("tuples_fetched", stat.IdxTupFetch),
			)
		}
	}
	
	return nil
}

// CheckSlowQueries identifies slow queries that might benefit from indexes
func (m *IndexMonitor) CheckSlowQueries(ctx context.Context, thresholdMs float64) error {
	// Check if pg_stat_statements is available
	var extensionExists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements'
		)
	`
	if err := m.db.WithContext(ctx).Raw(checkQuery).Scan(&extensionExists).Error; err != nil {
		return fmt.Errorf("failed to check pg_stat_statements extension: %w", err)
	}
	
	if !extensionExists {
		m.logger.Info("pg_stat_statements extension not available")
		return nil
	}
	
	query := `
		SELECT 
			query,
			calls,
			mean_exec_time,
			total_exec_time,
			stddev_exec_time
		FROM pg_stat_statements
		WHERE mean_exec_time > ?
		AND query NOT LIKE '%pg_stat_statements%'
		ORDER BY mean_exec_time DESC
		LIMIT 20
	`
	
	var slowQueries []SlowQuery
	if err := m.db.WithContext(ctx).Raw(query, thresholdMs).Scan(&slowQueries).Error; err != nil {
		m.logger.Debug("Could not fetch slow queries", zap.Error(err))
		return nil
	}
	
	for _, sq := range slowQueries {
		m.logger.Info("Slow query detected",
			zap.String("query", truncateQuery(sq.Query)),
			zap.Int64("calls", sq.Calls),
			zap.Float64("mean_time_ms", sq.MeanExecTime),
			zap.Float64("total_time_ms", sq.TotalExecTime),
			zap.Float64("stddev_ms", sq.StddevTime),
		)
	}
	
	return nil
}

// CheckTableStats checks table statistics and maintenance needs
func (m *IndexMonitor) CheckTableStats(ctx context.Context) error {
	query := `
		SELECT 
			schemaname || '.' || tablename as table_name,
			n_live_tup as total_rows,
			n_dead_tup as dead_rows,
			last_vacuum,
			last_autovacuum,
			idx_scan as index_scans,
			seq_scan as seq_scans
		FROM pg_stat_user_tables
		WHERE schemaname = 'public'
		ORDER BY n_live_tup DESC
	`
	
	var stats []TableStats
	if err := m.db.WithContext(ctx).Raw(query).Scan(&stats).Error; err != nil {
		return fmt.Errorf("failed to fetch table stats: %w", err)
	}
	
	for _, stat := range stats {
		logger := m.logger.With(
			zap.String("table", stat.TableName),
			zap.Int64("rows", stat.TotalRows),
			zap.Int64("dead_rows", stat.DeadRows),
		)
		
		// Check for high dead tuple ratio
		if stat.TotalRows > 0 {
			deadRatio := float64(stat.DeadRows) / float64(stat.TotalRows)
			if deadRatio > 0.2 {
				logger.Warn("High dead tuple ratio",
					zap.Float64("ratio", deadRatio),
				)
			}
		}
		
		// Check for tables with high sequential scan ratio
		if stat.IndexScans+stat.SeqScans > 0 {
			seqScanRatio := float64(stat.SeqScans) / float64(stat.IndexScans+stat.SeqScans)
			if seqScanRatio > 0.5 && stat.TotalRows > 1000 {
				logger.Warn("High sequential scan ratio",
					zap.Float64("ratio", seqScanRatio),
					zap.Int64("seq_scans", stat.SeqScans),
					zap.Int64("index_scans", stat.IndexScans),
				)
			}
		}
		
		// Check vacuum status
		if stat.LastAutoVacuum != nil {
			daysSinceVacuum := time.Since(*stat.LastAutoVacuum).Hours() / 24
			if daysSinceVacuum > 7 {
				logger.Warn("Table not vacuumed recently",
					zap.Float64("days_since_vacuum", daysSinceVacuum),
				)
			}
		}
	}
	
	return nil
}

// CheckIndexBloat checks for index bloat
func (m *IndexMonitor) CheckIndexBloat(ctx context.Context) error {
	query := `
		WITH index_bloat AS (
			SELECT
				schemaname,
				tablename,
				indexname,
				pg_relation_size(indexrelid) AS index_size,
				CASE WHEN indisunique THEN 1 ELSE 0 END AS is_unique,
				pg_stat_get_live_tuples(indrelid) AS num_rows
			FROM pg_stat_user_indexes
			JOIN pg_index ON pg_index.indexrelid = pg_stat_user_indexes.indexrelid
			WHERE schemaname = 'public'
		)
		SELECT 
			schemaname,
			tablename,
			indexname,
			pg_size_pretty(index_size) as index_size,
			CASE 
				WHEN num_rows > 0 THEN 
					ROUND(100.0 * index_size / (num_rows * 40), 2)
				ELSE 0 
			END as bloat_ratio
		FROM index_bloat
		WHERE index_size > 1024 * 1024 -- Only indexes > 1MB
		ORDER BY bloat_ratio DESC
	`
	
	type IndexBloat struct {
		SchemaName string
		TableName  string
		IndexName  string
		IndexSize  string
		BloatRatio float64
	}
	
	var results []IndexBloat
	if err := m.db.WithContext(ctx).Raw(query).Scan(&results).Error; err != nil {
		m.logger.Debug("Could not check index bloat", zap.Error(err))
		return nil
	}
	
	for _, result := range results {
		if result.BloatRatio > 200 {
			m.logger.Warn("Bloated index detected",
				zap.String("index", result.IndexName),
				zap.String("table", result.TableName),
				zap.String("size", result.IndexSize),
				zap.Float64("bloat_ratio", result.BloatRatio),
			)
		}
	}
	
	return nil
}

// RunMaintenance runs index maintenance tasks
func (m *IndexMonitor) RunMaintenance(ctx context.Context) error {
	m.logger.Info("Starting index maintenance")
	
	// Update statistics
	if err := m.updateStatistics(ctx); err != nil {
		m.logger.Error("Failed to update statistics", zap.Error(err))
	}
	
	// Check and reindex bloated indexes
	if err := m.reindexBloated(ctx); err != nil {
		m.logger.Error("Failed to reindex bloated indexes", zap.Error(err))
	}
	
	m.logger.Info("Index maintenance completed")
	return nil
}

// updateStatistics runs ANALYZE on all tables
func (m *IndexMonitor) updateStatistics(ctx context.Context) error {
	tables := []string{
		"clearing_operations",
		"clearing_tokens",
		"payment_records",
		"refunds",
	}
	
	for _, table := range tables {
		query := fmt.Sprintf("ANALYZE %s", table)
		if err := m.db.WithContext(ctx).Exec(query).Error; err != nil {
			m.logger.Error("Failed to analyze table",
				zap.String("table", table),
				zap.Error(err),
			)
			continue
		}
		m.logger.Debug("Updated statistics for table", zap.String("table", table))
	}
	
	return nil
}

// reindexBloated reindexes heavily bloated indexes
func (m *IndexMonitor) reindexBloated(ctx context.Context) error {
	// Check for bloated indexes
	query := `
		SELECT 
			indexname,
			tablename
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
		AND pg_relation_size(indexrelid) > 10 * 1024 * 1024 -- > 10MB
	`
	
	type IndexInfo struct {
		IndexName string
		TableName string
	}
	
	var indexes []IndexInfo
	if err := m.db.WithContext(ctx).Raw(query).Scan(&indexes).Error; err != nil {
		return fmt.Errorf("failed to get index list: %w", err)
	}
	
	for _, idx := range indexes {
		// Use REINDEX CONCURRENTLY to avoid locking
		reindexQuery := fmt.Sprintf("REINDEX INDEX CONCURRENTLY %s", idx.IndexName)
		
		m.logger.Info("Reindexing bloated index",
			zap.String("index", idx.IndexName),
			zap.String("table", idx.TableName),
		)
		
		if err := m.db.WithContext(ctx).Exec(reindexQuery).Error; err != nil {
			m.logger.Error("Failed to reindex",
				zap.String("index", idx.IndexName),
				zap.Error(err),
			)
			continue
		}
	}
	
	return nil
}

// StartMonitoring starts periodic monitoring
func (m *IndexMonitor) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.logger.Debug("Running index monitoring")
			
			if err := m.CheckIndexUsage(ctx); err != nil {
				m.logger.Error("Index usage check failed", zap.Error(err))
			}
			
			if err := m.CheckSlowQueries(ctx, 100); err != nil {
				m.logger.Error("Slow query check failed", zap.Error(err))
			}
			
			if err := m.CheckTableStats(ctx); err != nil {
				m.logger.Error("Table stats check failed", zap.Error(err))
			}
			
			if err := m.CheckIndexBloat(ctx); err != nil {
				m.logger.Error("Index bloat check failed", zap.Error(err))
			}
		}
	}
}

// truncateQuery truncates long queries for logging
func truncateQuery(query string) string {
	const maxLength = 200
	if len(query) <= maxLength {
		return query
	}
	return query[:maxLength] + "..."
}