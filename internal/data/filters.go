package data

import (
	"math"
	"strings"

	"greenlight.kerseeehuang.com/internal/validator"
)

type Filters struct {
	Page         int      // The page number of results
	PageSize     int      // Size of one page
	Sort         string   // Name of field by which returned records are sorted
	SortSafelist []string // List of field name
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// ValidateFilters validates Filters f and store the validation error into v.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "must be less than 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be less than 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}

// sortColumn() checks that the client-provided sort field matches f.SortSafelist.
// If it matches, return the sort field with the tripped leading hyphen.
// If it does not match, then panic.
func (f Filters) sortColumn() string {
	for _, safeVal := range f.SortSafelist {
		if f.Sort == safeVal {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection return "ASC" or "DESC" depending on f.Sort.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

// limit() returns f.PageSize.
func (f Filters) limit() int {
	return f.PageSize
}

// offset returns the offset of record in table movies in DB.
// offset = (f.Page - 1) * f.PageSize
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// calculateMatadata return a Metadata whose fields are
// calculated by totalRecords, page and pageSize.
func calculateMatadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
