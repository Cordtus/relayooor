package database

import (
	"fmt"
	"strings"
	"time"
)

// QueryBuilder helps construct optimized SQL queries
type QueryBuilder struct {
	table      string
	selectCols []string
	joins      []string
	wheres     []string
	args       []interface{}
	groupBy    []string
	orderBy    []string
	limit      int
	offset     int
	forUpdate  bool
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		selectCols: []string{"*"},
		args:       make([]interface{}, 0),
	}
}

// Select specifies columns to select
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(joinType, table, on string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("%s JOIN %s ON %s", joinType, table, on))
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.wheres = append(qb.wheres, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = fmt.Sprintf("$%d", len(qb.args)+i+1)
		qb.args = append(qb.args, values[i])
	}
	
	qb.wheres = append(qb.wheres, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
	return qb
}

// WhereBetween adds a WHERE BETWEEN condition
func (qb *QueryBuilder) WhereBetween(column string, start, end interface{}) *QueryBuilder {
	condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", column, len(qb.args)+1, len(qb.args)+2)
	qb.wheres = append(qb.wheres, condition)
	qb.args = append(qb.args, start, end)
	return qb
}

// GroupBy adds GROUP BY columns
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = columns
	return qb
}

// OrderBy adds ORDER BY columns
func (qb *QueryBuilder) OrderBy(column string, desc bool) *QueryBuilder {
	order := column
	if desc {
		order += " DESC"
	} else {
		order += " ASC"
	}
	qb.orderBy = append(qb.orderBy, order)
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// ForUpdate adds FOR UPDATE clause
func (qb *QueryBuilder) ForUpdate() *QueryBuilder {
	qb.forUpdate = true
	return qb
}

// Build constructs the final SQL query
func (qb *QueryBuilder) Build() (string, []interface{}) {
	parts := []string{
		"SELECT",
		strings.Join(qb.selectCols, ", "),
		"FROM",
		qb.table,
	}

	// Add joins
	if len(qb.joins) > 0 {
		parts = append(parts, strings.Join(qb.joins, " "))
	}

	// Add where conditions
	if len(qb.wheres) > 0 {
		parts = append(parts, "WHERE", strings.Join(qb.wheres, " AND "))
	}

	// Add group by
	if len(qb.groupBy) > 0 {
		parts = append(parts, "GROUP BY", strings.Join(qb.groupBy, ", "))
	}

	// Add order by
	if len(qb.orderBy) > 0 {
		parts = append(parts, "ORDER BY", strings.Join(qb.orderBy, ", "))
	}

	// Add limit
	if qb.limit > 0 {
		parts = append(parts, fmt.Sprintf("LIMIT %d", qb.limit))
	}

	// Add offset
	if qb.offset > 0 {
		parts = append(parts, fmt.Sprintf("OFFSET %d", qb.offset))
	}

	// Add for update
	if qb.forUpdate {
		parts = append(parts, "FOR UPDATE")
	}

	return strings.Join(parts, " "), qb.args
}

// BatchInsertBuilder helps construct batch insert queries
type BatchInsertBuilder struct {
	table      string
	columns    []string
	values     [][]interface{}
	onConflict string
	returning  string
}

// NewBatchInsertBuilder creates a new batch insert builder
func NewBatchInsertBuilder(table string, columns ...string) *BatchInsertBuilder {
	return &BatchInsertBuilder{
		table:   table,
		columns: columns,
		values:  make([][]interface{}, 0),
	}
}

// AddRow adds a row to insert
func (b *BatchInsertBuilder) AddRow(values ...interface{}) *BatchInsertBuilder {
	if len(values) != len(b.columns) {
		panic("number of values must match number of columns")
	}
	b.values = append(b.values, values)
	return b
}

// OnConflict sets the ON CONFLICT clause
func (b *BatchInsertBuilder) OnConflict(clause string) *BatchInsertBuilder {
	b.onConflict = clause
	return b
}

// Returning sets the RETURNING clause
func (b *BatchInsertBuilder) Returning(clause string) *BatchInsertBuilder {
	b.returning = clause
	return b
}

// Build constructs the final batch insert query
func (b *BatchInsertBuilder) Build() (string, []interface{}) {
	if len(b.values) == 0 {
		return "", nil
	}

	// Build query parts
	parts := []string{
		"INSERT INTO",
		b.table,
		"(" + strings.Join(b.columns, ", ") + ")",
		"VALUES",
	}

	// Build value placeholders
	args := make([]interface{}, 0, len(b.values)*len(b.columns))
	valueParts := make([]string, len(b.values))
	
	argIndex := 1
	for i, row := range b.values {
		placeholders := make([]string, len(row))
		for j, val := range row {
			placeholders[j] = fmt.Sprintf("$%d", argIndex)
			args = append(args, val)
			argIndex++
		}
		valueParts[i] = "(" + strings.Join(placeholders, ", ") + ")"
	}
	
	parts = append(parts, strings.Join(valueParts, ", "))

	// Add ON CONFLICT clause
	if b.onConflict != "" {
		parts = append(parts, "ON CONFLICT", b.onConflict)
	}

	// Add RETURNING clause
	if b.returning != "" {
		parts = append(parts, "RETURNING", b.returning)
	}

	return strings.Join(parts, " "), args
}

// QueryCache provides simple query result caching
type QueryCache struct {
	cache map[string]cachedResult
	ttl   time.Duration
}

type cachedResult struct {
	data      interface{}
	expiresAt time.Time
}

// NewQueryCache creates a new query cache
func NewQueryCache(ttl time.Duration) *QueryCache {
	return &QueryCache{
		cache: make(map[string]cachedResult),
		ttl:   ttl,
	}
}

// Get retrieves a cached result
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	result, exists := qc.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(result.expiresAt) {
		delete(qc.cache, key)
		return nil, false
	}

	return result.data, true
}

// Set stores a result in cache
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.cache[key] = cachedResult{
		data:      data,
		expiresAt: time.Now().Add(qc.ttl),
	}
}

// Clear removes all cached results
func (qc *QueryCache) Clear() {
	qc.cache = make(map[string]cachedResult)
}

// Cleanup removes expired entries
func (qc *QueryCache) Cleanup() {
	now := time.Now()
	for key, result := range qc.cache {
		if now.After(result.expiresAt) {
			delete(qc.cache, key)
		}
	}
}