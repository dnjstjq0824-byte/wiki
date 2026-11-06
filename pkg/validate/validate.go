// Package validate provides shared HTTP input validation, pagination helpers,
// and a consistent JSON error writer for Treewiki handlers.
package validate

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// Error writes a JSON error body matching the gateway-wide {"error": msg} shape.
func Error(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// NonEmpty ensures a required string field is present.
func NonEmpty(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(field + " is required")
	}
	return nil
}

// MaxLen ensures value does not exceed max runes.
func MaxLen(field, value string, max int) error {
	if len([]rune(value)) > max {
		return errors.New(field + " exceeds maximum length")
	}
	return nil
}

// OneOf ensures value is one of the allowed set (case-sensitive).
func OneOf(field, value string, allowed ...string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return errors.New(field + " has invalid value")
}

// Pagination carries validated page/per_page query parameters.
type Pagination struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// ParsePagination reads ?page=&per_page=, applying defaults and hard clamps.
// Default page=1, perPage=20, max perPage=100.
func ParsePagination(r *http.Request, defaultPerPage int) Pagination {
	if defaultPerPage <= 0 {
		defaultPerPage = 20
	}
	if defaultPerPage > 100 {
		defaultPerPage = 100
	}
	p := Pagination{Page: 1, PerPage: defaultPerPage}
	if v := strings.TrimSpace(r.URL.Query().Get("page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Page = n
		}
	}
	if v := strings.TrimSpace(r.URL.Query().Get("per_page")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.PerPage = n
		}
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	return p
}

// Offset returns the slice offset for this page.
func (p Pagination) Offset() int { return (p.Page - 1) * p.PerPage }

// HasPagination reports whether the request asks for a paginated envelope.
func HasPagination(r *http.Request) bool {
	q := r.URL.Query()
	return q.Has("page") || q.Has("per_page")
}

// Paginated is the standard envelope for list endpoints.
type Paginated[T any] struct {
	Data    []T `json:"data"`
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Paginate slices items for the requested page and wraps them in an envelope.
// A nil items slice yields an empty (non-nil) data array.
func Paginate[T any](items []T, p Pagination) Paginated[T] {
	if items == nil {
		items = []T{}
	}
	total := len(items)
	start := p.Offset()
	if start >= total {
		return Paginated[T]{Data: []T{}, Total: total, Page: p.Page, PerPage: p.PerPage}
	}
	end := start + p.PerPage
	if end > total {
		end = total
	}
	return Paginated[T]{Data: items[start:end], Total: total, Page: p.Page, PerPage: p.PerPage}
}
