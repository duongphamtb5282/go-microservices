package models

import (
	"math"
)

// PaginationRequest represents a pagination request
type PaginationRequest struct {
	Page    int    `json:"page" form:"page" validate:"min=1"`
	Limit   int    `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy  string `json:"sort_by" form:"sort_by"`
	SortDir string `json:"sort_dir" form:"sort_dir" validate:"oneof=asc desc"`
}

// NewPaginationRequest creates a new pagination request
func NewPaginationRequest(page, limit int) PaginationRequest {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return PaginationRequest{
		Page:    page,
		Limit:   limit,
		SortBy:  "created_at",
		SortDir: "desc",
	}
}

// GetOffset returns the offset for the pagination
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// GetSortBy returns the sort by field
func (p *PaginationRequest) GetSortBy() string {
	if p.SortBy == "" {
		return "created_at"
	}
	return p.SortBy
}

// GetSortDir returns the sort direction
func (p *PaginationRequest) GetSortDir() string {
	if p.SortDir == "" {
		return "desc"
	}
	return p.SortDir
}

// PaginationResponse represents a pagination response
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// NewPaginationResponse creates a new pagination response
func NewPaginationResponse(page, limit int, total int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetNextPage returns the next page number
func (p *PaginationResponse) GetNextPage() int {
	if p.HasNext {
		return p.Page + 1
	}
	return p.Page
}

// GetPrevPage returns the previous page number
func (p *PaginationResponse) GetPrevPage() int {
	if p.HasPrev {
		return p.Page - 1
	}
	return p.Page
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, page, limit int, total int64) PaginatedResponse {
	return PaginatedResponse{
		Data:       data,
		Pagination: NewPaginationResponse(page, limit, total),
	}
}

// CursorPaginationRequest represents a cursor-based pagination request
type CursorPaginationRequest struct {
	Cursor  string `json:"cursor" form:"cursor"`
	Limit   int    `json:"limit" form:"limit" validate:"min=1,max=100"`
	SortBy  string `json:"sort_by" form:"sort_by"`
	SortDir string `json:"sort_dir" form:"sort_dir" validate:"oneof=asc desc"`
}

// NewCursorPaginationRequest creates a new cursor pagination request
func NewCursorPaginationRequest(cursor string, limit int) CursorPaginationRequest {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return CursorPaginationRequest{
		Cursor:  cursor,
		Limit:   limit,
		SortBy:  "created_at",
		SortDir: "desc",
	}
}

// GetSortBy returns the sort by field
func (c *CursorPaginationRequest) GetSortBy() string {
	if c.SortBy == "" {
		return "created_at"
	}
	return c.SortBy
}

// GetSortDir returns the sort direction
func (c *CursorPaginationRequest) GetSortDir() string {
	if c.SortDir == "" {
		return "desc"
	}
	return c.SortDir
}

// CursorPaginationResponse represents a cursor-based pagination response
type CursorPaginationResponse struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasNext    bool        `json:"has_next"`
}

// NewCursorPaginationResponse creates a new cursor pagination response
func NewCursorPaginationResponse(data interface{}, nextCursor string, hasNext bool) CursorPaginationResponse {
	return CursorPaginationResponse{
		Data:       data,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query      string                 `json:"query" form:"query"`
	Fields     []string               `json:"fields" form:"fields"`
	Filters    map[string]interface{} `json:"filters" form:"filters"`
	Pagination PaginationRequest      `json:"pagination"`
}

// NewSearchRequest creates a new search request
func NewSearchRequest(query string, page, limit int) SearchRequest {
	return SearchRequest{
		Query:      query,
		Fields:     make([]string, 0),
		Filters:    make(map[string]interface{}),
		Pagination: NewPaginationRequest(page, limit),
	}
}

// AddField adds a field to search
func (s *SearchRequest) AddField(field string) {
	s.Fields = append(s.Fields, field)
}

// AddFilter adds a filter
func (s *SearchRequest) AddFilter(key string, value interface{}) {
	if s.Filters == nil {
		s.Filters = make(map[string]interface{})
	}
	s.Filters[key] = value
}

// SearchResponse represents a search response
type SearchResponse struct {
	Data       interface{}            `json:"data"`
	Query      string                 `json:"query"`
	Total      int64                  `json:"total"`
	Pagination PaginationResponse     `json:"pagination"`
	Facets     map[string]interface{} `json:"facets,omitempty"`
}

// NewSearchResponse creates a new search response
func NewSearchResponse(data interface{}, query string, total int64, page, limit int) SearchResponse {
	return SearchResponse{
		Data:       data,
		Query:      query,
		Total:      total,
		Pagination: NewPaginationResponse(page, limit, total),
		Facets:     make(map[string]interface{}),
	}
}

// AddFacet adds a facet to the search response
func (s *SearchResponse) AddFacet(key string, value interface{}) {
	if s.Facets == nil {
		s.Facets = make(map[string]interface{})
	}
	s.Facets[key] = value
}
