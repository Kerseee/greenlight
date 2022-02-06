package data

import "greenlight.kerseeehuang.com/internal/validator"

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

// ValidateFilters validates Filters f and store the validation error into v.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than 0")
	v.Check(f.Page <= 10_000_000, "page", "must be less than 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than 0")
	v.Check(f.PageSize <= 100, "page_size", "must be less than 100")
	v.Check(validator.In(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
