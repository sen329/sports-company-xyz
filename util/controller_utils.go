package util

import (
	"math"
)

// PaginationParams defines the structure for pagination query parameters.
type PaginationParams struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// SetPaginationDefaults validates and sets default values for page and limit.
func SetPaginationDefaults(page, limit int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10 // Default page size
	}
	return page, limit
}

// CalculateTotalPages computes the total number of pages for a paginated response.
func CalculateTotalPages(totalRecords int64, limit int) int {
	if totalRecords == 0 {
		return 0
	}
	return int(math.Ceil(float64(totalRecords) / float64(limit)))
}
