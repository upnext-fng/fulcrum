package store

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// ==========================================
// PAGINATION STRUCTURES
// ==========================================

// PaginationParams represents pagination parameters from request
type PaginationParams struct {
	Page   int    `json:"page" form:"page" validate:"min=1"`
	Limit  int    `json:"limit" form:"limit" validate:"min=1,max=100"`
	Sort   string `json:"sort" form:"sort"`
	Order  string `json:"order" form:"order" validate:"oneof=asc desc"`
	Search string `json:"search" form:"search"`
	Filter string `json:"filter" form:"filter"`
}

// PaginationResult represents paginated query results with metadata
type PaginationResult struct {
	Data       any             `json:"data"`
	Pagination *PaginationMeta `json:"pagination"`
	Sort       *SortMeta       `json:"sort,omitempty"`
	Search     *SearchMeta     `json:"search,omitempty"`
	Filter     *FilterMeta     `json:"filter,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	Total        int64 `json:"total"`
	TotalPages   int   `json:"total_pages"`
	HasNext      bool  `json:"has_next"`
	HasPrevious  bool  `json:"has_previous"`
	NextPage     *int  `json:"next_page,omitempty"`
	PreviousPage *int  `json:"previous_page,omitempty"`
	Offset       int   `json:"offset"`
}

// SortMeta contains sorting metadata
type SortMeta struct {
	Field string `json:"field"`
	Order string `json:"order"`
}

// SearchMeta contains search metadata
type SearchMeta struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields,omitempty"`
}

// FilterMeta contains filter metadata
type FilterMeta struct {
	Applied bool           `json:"applied"`
	Filters map[string]any `json:"filters,omitempty"`
	Count   int            `json:"count,omitempty"`
}

// ==========================================
// PAGINATION CONFIGURATION
// ==========================================

// PaginationConfig represents pagination configuration
type PaginationConfig struct {
	DefaultPage       int      `mapstructure:"default_page"`
	DefaultLimit      int      `mapstructure:"default_limit"`
	MaxLimit          int      `mapstructure:"max_limit"`
	DefaultSort       string   `mapstructure:"default_sort"`
	DefaultOrder      string   `mapstructure:"default_order"`
	AllowedSortFields []string `mapstructure:"allowed_sort_fields"`
	SearchFields      []string `mapstructure:"search_fields"`
}

// DefaultPaginationConfig returns default pagination configuration
func DefaultPaginationConfig() *PaginationConfig {
	return &PaginationConfig{
		DefaultPage:       1,
		DefaultLimit:      20,
		MaxLimit:          100,
		DefaultSort:       "created_at",
		DefaultOrder:      "desc",
		AllowedSortFields: []string{"id", "created_at", "updated_at"},
		SearchFields:      []string{"name", "email"},
	}
}

// ==========================================
// PAGINATION PARAMS METHODS
// ==========================================

// NewPaginationParams creates new pagination parameters with defaults
func NewPaginationParams() *PaginationParams {
	return &PaginationParams{
		Page:  1,
		Limit: 20,
		Order: "desc",
	}
}

// DefaultPaginationParams returns default pagination parameters
func DefaultPaginationParams() *PaginationParams {
	return &PaginationParams{
		Page:  1,
		Limit: 20,
		Sort:  "created_at",
		Order: "desc",
	}
}

// WithDefaults applies default values to pagination parameters
func (p *PaginationParams) WithDefaults(config *PaginationConfig) *PaginationParams {
	if config == nil {
		config = DefaultPaginationConfig()
	}

	params := &PaginationParams{
		Page:   p.Page,
		Limit:  p.Limit,
		Sort:   p.Sort,
		Order:  p.Order,
		Search: p.Search,
		Filter: p.Filter,
	}

	if params.Page <= 0 {
		params.Page = config.DefaultPage
	}

	if params.Limit <= 0 {
		params.Limit = config.DefaultLimit
	}

	if params.Limit > config.MaxLimit {
		params.Limit = config.MaxLimit
	}

	if params.Sort == "" {
		params.Sort = config.DefaultSort
	}

	if params.Order == "" {
		params.Order = config.DefaultOrder
	}

	return params
}

// Validate validates pagination parameters
func (p *PaginationParams) Validate() error {
	if p.Page < 1 {
		return fmt.Errorf("page must be greater than 0, got %d", p.Page)
	}

	if p.Limit < 1 {
		return fmt.Errorf("limit must be greater than 0, got %d", p.Limit)
	}

	if p.Limit > 100 {
		return fmt.Errorf("limit must not exceed 100, got %d", p.Limit)
	}

	if p.Order != "" && p.Order != "asc" && p.Order != "desc" {
		return fmt.Errorf("order must be 'asc' or 'desc', got '%s'", p.Order)
	}

	return nil
}

// ValidateWithConfig validates pagination parameters against configuration
func (p *PaginationParams) ValidateWithConfig(config *PaginationConfig) error {
	if err := p.Validate(); err != nil {
		return err
	}

	if config == nil {
		return nil
	}

	if p.Limit > config.MaxLimit {
		return fmt.Errorf("limit exceeds maximum allowed (%d), got %d", config.MaxLimit, p.Limit)
	}

	if p.Sort != "" && len(config.AllowedSortFields) > 0 {
		allowed := false
		for _, field := range config.AllowedSortFields {
			if field == p.Sort {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("sort field '%s' is not allowed", p.Sort)
		}
	}

	return nil
}

// Offset calculates the offset for database queries
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// HasSort checks if sorting is specified
func (p *PaginationParams) HasSort() bool {
	return p.Sort != ""
}

// HasSearch checks if search is specified
func (p *PaginationParams) HasSearch() bool {
	return p.Search != ""
}

// HasFilter checks if filter is specified
func (p *PaginationParams) HasFilter() bool {
	return p.Filter != ""
}

// GetOrderClause returns the ORDER BY clause for SQL
func (p *PaginationParams) GetOrderClause() string {
	if !p.HasSort() {
		return ""
	}

	order := "ASC"
	if p.Order == "desc" {
		order = "DESC"
	}

	return fmt.Sprintf("%s %s", p.Sort, order)
}

// Clone creates a copy of pagination parameters
func (p *PaginationParams) Clone() *PaginationParams {
	return &PaginationParams{
		Page:   p.Page,
		Limit:  p.Limit,
		Sort:   p.Sort,
		Order:  p.Order,
		Search: p.Search,
		Filter: p.Filter,
	}
}

// ==========================================
// PAGINATION RESULT METHODS
// ==========================================

// NewPaginationResult creates a new pagination result
func NewPaginationResult(data any, params *PaginationParams, total int64) *PaginationResult {
	if params == nil {
		params = DefaultPaginationParams()
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))
	hasNext := params.Page < totalPages
	hasPrevious := params.Page > 1

	var nextPage, previousPage *int
	if hasNext {
		next := params.Page + 1
		nextPage = &next
	}
	if hasPrevious {
		prev := params.Page - 1
		previousPage = &prev
	}

	result := &PaginationResult{
		Data: data,
		Pagination: &PaginationMeta{
			Page:         params.Page,
			Limit:        params.Limit,
			Total:        total,
			TotalPages:   totalPages,
			HasNext:      hasNext,
			HasPrevious:  hasPrevious,
			NextPage:     nextPage,
			PreviousPage: previousPage,
			Offset:       params.Offset(),
		},
	}

	// Add sort metadata if sorting is applied
	if params.HasSort() {
		result.Sort = &SortMeta{
			Field: params.Sort,
			Order: params.Order,
		}
	}

	// Add search metadata if search is applied
	if params.HasSearch() {
		result.Search = &SearchMeta{
			Query: params.Search,
		}
	}

	// Add filter metadata if filter is applied
	if params.HasFilter() {
		result.Filter = &FilterMeta{
			Applied: true,
		}
	}

	return result
}

// IsEmpty checks if the result contains no data
func (r *PaginationResult) IsEmpty() bool {
	return r.Pagination.Total == 0
}

// IsFirstPage checks if this is the first page
func (r *PaginationResult) IsFirstPage() bool {
	return r.Pagination.Page == 1
}

// IsLastPage checks if this is the last page
func (r *PaginationResult) IsLastPage() bool {
	return r.Pagination.Page == r.Pagination.TotalPages
}

// GetCurrentPageSize returns the number of items in current page
func (r *PaginationResult) GetCurrentPageSize() int {
	switch data := r.Data.(type) {
	case []any:
		return len(data)
	case []*any:
		return len(data)
	default:
		// Use reflection to get slice length if needed
		return 0
	}
}

// ==========================================
// PAGINATION UTILITIES
// ==========================================

// Paginate applies pagination to a GORM query and returns results
func Paginate(ctx context.Context, db *gorm.DB, params *PaginationParams, dest any) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	// Validate parameters
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pagination parameters: %w", err)
	}

	// Count total records
	var total int64
	countQuery := db.Session(&gorm.Session{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Apply pagination and sorting
	query := db.Offset(params.Offset()).Limit(params.Limit)

	// Apply sorting if specified
	if params.HasSort() {
		query = query.Order(params.GetOrderClause())
	}

	// Execute query
	if err := query.Find(dest).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch paginated records: %w", err)
	}

	return NewPaginationResult(dest, params, total), nil
}

// PaginateWithSearch applies pagination and search to a GORM query
func PaginateWithSearch(ctx context.Context, db *gorm.DB, params *PaginationParams, searchFields []string, dest any) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	query := db

	// Apply search if specified
	if params.HasSearch() && len(searchFields) > 0 {
		searchQuery := db.Where("1=0") // Start with false condition

		for _, field := range searchFields {
			searchQuery = searchQuery.Or(fmt.Sprintf("%s ILIKE ?", field), "%"+params.Search+"%")
		}

		query = query.Where(searchQuery)
	}

	result, err := Paginate(ctx, query, params, dest)
	if err != nil {
		return nil, err
	}

	// Add search metadata
	if params.HasSearch() {
		result.Search = &SearchMeta{
			Query:  params.Search,
			Fields: searchFields,
		}
	}

	return result, nil
}

// PaginateWithCondition applies pagination with a generic condition
func PaginateWithCondition[T Entity](ctx context.Context, db *gorm.DB, params *PaginationParams, condition Condition[T], dest any) (*PaginationResult, error) {
	if params == nil {
		params = DefaultPaginationParams()
	}

	query := db

	// Apply condition if provided
	if condition != nil {
		if err := condition.Validate(); err != nil {
			return nil, fmt.Errorf("invalid condition: %w", err)
		}

		sql, args := condition.ToSQL()
		if sql != "" {
			query = query.Where(sql, args...)
		}
	}

	return Paginate(ctx, query, params, dest)
}

// ==========================================
// HELPER FUNCTIONS
// ==========================================

// CalculateTotalPages calculates total pages based on total records and limit
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}

// CalculateOffset calculates offset based on page and limit
func CalculateOffset(page, limit int) int {
	if page <= 0 {
		page = 1
	}
	return (page - 1) * limit
}

// ValidatePageRange validates if page is within valid range
func ValidatePageRange(page int, totalPages int) error {
	if page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}

	if totalPages > 0 && page > totalPages {
		return fmt.Errorf("page %d exceeds total pages %d", page, totalPages)
	}

	return nil
}

// ==========================================
// PAGINATION BUILDER
// ==========================================

// PaginationBuilder provides a fluent interface for building pagination
type PaginationBuilder struct {
	params *PaginationParams
	config *PaginationConfig
}

// NewPaginationBuilder creates a new pagination builder
func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{
		params: NewPaginationParams(),
		config: DefaultPaginationConfig(),
	}
}

// Page sets the page number
func (b *PaginationBuilder) Page(page int) *PaginationBuilder {
	b.params.Page = page
	return b
}

// Limit sets the limit
func (b *PaginationBuilder) Limit(limit int) *PaginationBuilder {
	b.params.Limit = limit
	return b
}

// Sort sets the sort field
func (b *PaginationBuilder) Sort(field string) *PaginationBuilder {
	b.params.Sort = field
	return b
}

// Order sets the sort order
func (b *PaginationBuilder) Order(order string) *PaginationBuilder {
	b.params.Order = order
	return b
}

// Search sets the search query
func (b *PaginationBuilder) Search(query string) *PaginationBuilder {
	b.params.Search = query
	return b
}

// Filter sets the filter
func (b *PaginationBuilder) Filter(filter string) *PaginationBuilder {
	b.params.Filter = filter
	return b
}

// WithConfig sets the pagination configuration
func (b *PaginationBuilder) WithConfig(config *PaginationConfig) *PaginationBuilder {
	b.config = config
	return b
}

// Build builds the final pagination parameters
func (b *PaginationBuilder) Build() (*PaginationParams, error) {
	params := b.params.WithDefaults(b.config)

	if err := params.ValidateWithConfig(b.config); err != nil {
		return nil, err
	}

	return params, nil
}
