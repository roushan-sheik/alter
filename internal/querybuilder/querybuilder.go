// Package querybuilder provides a fluent, chainable query builder utility
// for GORM-based PostgreSQL repositories, inspired by the Repository Pattern.
//
// It supports search (ILIKE), pagination, sorting, and arbitrary column filtering
// from a flat map of query parameters (e.g. parsed from HTTP query strings).
//
// Usage:
//
//	params := querybuilder.Params{
//	    "search": "morning",
//	    "page":   "2",
//	    "limit":  "5",
//	    "sort":   "-created_at",
//	}
//
//	var results []*motivation.Motivation
//	meta, err := querybuilder.New(db, params).
//	    Search([]string{"title", "speaker_name"}).
//	    Filter().
//	    Sort().
//	    Paginate().
//	    Execute(&results)
package querybuilder

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Params is a flat map of query parameters, typically parsed from an HTTP request.
// Example: url.Values converted via a helper, or a custom map[string]string.
type Params map[string]string

// Meta carries pagination metadata returned alongside a paginated result set.
type Meta struct {
	// Page is the current page number (1-indexed).
	Page int `json:"page"`
	// Limit is the number of items per page.
	Limit int `json:"limit"`
	// Total is the total number of matching records.
	Total int64 `json:"total"`
	// TotalPages is the total number of pages.
	TotalPages int `json:"total_pages"`
}

// reserved keys that are consumed by the builder and MUST NOT be passed
// as column filters to the database.
var reservedKeys = map[string]bool{
	"search": true,
	"page":   true,
	"limit":  true,
	"sort":   true,
	"fields": true,
}

const (
	defaultPage  = 1
	defaultLimit = 10
	defaultSort  = "created_at desc"
)

// QueryBuilder is a fluent builder that wraps a *gorm.DB and a Params map.
// Chain methods to progressively apply clauses, then call Execute to run the query.
type QueryBuilder struct {
	db     *gorm.DB
	params Params
}

// New creates a new QueryBuilder.
//
//	db     – a *gorm.DB scoped to the relevant table (e.g. db.Model(&Motivation{}))
//	params – flat string map of query parameters
func New(db *gorm.DB, params Params) *QueryBuilder {
	return &QueryBuilder{db: db, params: params}
}

// Search applies a case-insensitive ILIKE search across the given columns.
// All columns are OR-ed together.
//
//	.Search([]string{"title", "speaker_name"})
//	→ WHERE (title ILIKE '%q%' OR speaker_name ILIKE '%q%')
//
// The search term is read from params["search"].
// If the key is missing or empty the clause is skipped.
func (qb *QueryBuilder) Search(columns []string) *QueryBuilder {
	term := strings.TrimSpace(qb.params["search"])
	if term == "" || len(columns) == 0 {
		return qb
	}

	pattern := "%" + term + "%"

	// Build: "col1 ILIKE ? OR col2 ILIKE ? …"
	clauses := make([]string, len(columns))
	args := make([]interface{}, len(columns))
	for i, col := range columns {
		clauses[i] = fmt.Sprintf("%s ILIKE ?", col)
		args[i] = pattern
	}

	qb.db = qb.db.Where(strings.Join(clauses, " OR "), args...)
	return qb
}

// Filter applies equality filters for any params key that is NOT in the
// reserved key set (search, page, limit, sort, fields).
//
// Example:  params["speaker_name"] = "John" → WHERE speaker_name = 'John'
//
// Only non-empty values are applied.
func (qb *QueryBuilder) Filter() *QueryBuilder {
	for key, val := range qb.params {
		if reservedKeys[key] || strings.TrimSpace(val) == "" {
			continue
		}
		qb.db = qb.db.Where(fmt.Sprintf("%s = ?", key), val)
	}
	return qb
}

// Sort applies an ORDER BY clause.
// The sort value is read from params["sort"].
//
// Supported formats:
//   - "created_at"  → ORDER BY created_at ASC
//   - "-created_at" → ORDER BY created_at DESC  (leading minus = descending)
//
// Defaults to "created_at desc" when the param is absent or empty.
func (qb *QueryBuilder) Sort() *QueryBuilder {
	raw := strings.TrimSpace(qb.params["sort"])
	if raw == "" {
		qb.db = qb.db.Order(defaultSort)
		return qb
	}

	if strings.HasPrefix(raw, "-") {
		col := raw[1:]
		qb.db = qb.db.Order(fmt.Sprintf("%s desc", col))
	} else {
		qb.db = qb.db.Order(fmt.Sprintf("%s asc", raw))
	}
	return qb
}

// Paginate applies LIMIT / OFFSET clauses based on params["page"] and params["limit"].
// Must be called BEFORE Execute so that the separate count query runs on the
// un-paginated scope.
func (qb *QueryBuilder) Paginate() *QueryBuilder {
	limit := parseIntParam(qb.params["limit"], defaultLimit)
	page := parseIntParam(qb.params["page"], defaultPage)
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}

	offset := (page - 1) * limit
	qb.db = qb.db.Limit(limit).Offset(offset)
	return qb
}

// Execute runs the assembled query, scanning results into dest.
// dest must be a pointer to a slice, e.g. *[]*Motivation.
//
// It returns a populated Meta struct (total count, pages, etc.).
// Execute does NOT run a COUNT query internally – call ExecuteWithMeta for that.
func (qb *QueryBuilder) Execute(dest interface{}) error {
	return qb.db.Find(dest).Error
}

// ExecuteWithMeta runs TWO queries:
//  1. A COUNT(*) on the current scope (before LIMIT/OFFSET) to populate Meta.
//  2. The full paginated query to populate dest.
//
// dest must be a pointer to a slice, e.g. *[]*Motivation.
//
// Example:
//
//	var items []*Motivation
//	meta, err := qb.Search(cols).Filter().Sort().Paginate().ExecuteWithMeta(&items)
func (qb *QueryBuilder) ExecuteWithMeta(dest interface{}) (*Meta, error) {
	limit := parseIntParam(qb.params["limit"], defaultLimit)
	page := parseIntParam(qb.params["page"], defaultPage)
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}

	// Count query: clone the session WITHOUT limit/offset.
	var total int64
	countDB := qb.db.Session(&gorm.Session{}).Limit(-1).Offset(-1)
	if err := countDB.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("querybuilder count: %w", err)
	}

	// Data query.
	if err := qb.db.Find(dest).Error; err != nil {
		return nil, fmt.Errorf("querybuilder find: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &Meta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

//  helpers

func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
