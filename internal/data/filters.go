package data

import (
	"math"
	"strings"

	"greenlight.ricci2511.dev/internal/validator"
)

type Filters struct {
	Page         int      // Pagination page number
	PageSize     int      // Number of records per page
	Sort         string   // Field name to sort by
	SortSafeList []string // List of allowed field names to sort by
}

// Runs validation checks on the filter parameters provided by the client.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10,000,000")
	v.Check(f.PageSize > 0, "pageSize", "must be greater than zero")
	v.Check(f.PageSize <= 100, "pageSize", "must be a maximum of 100")
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

// Returns the column name to sort by based on the Sort field value provided by the client.
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	// This should never happen if the ValidateFilters() function above is used.
	// This is just a failsafe.
	panic("unsafe sort parameter: " + f.Sort)
}

// Returns the sort direction based on the prefix character of the Sort field value.
//
// Hyphen prefix (-) indicates descending order, otherwise ascending.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

// Returns the SQL LIMIT clause value based on the PageSize field value provided by the client.
func (f Filters) limit() int {
	return f.PageSize
}

// Returns the SQL OFFSET clause value based on the Page and PageSize field values provided by the client.
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

type Metadata struct {
	CurrentPage  int `json:"currentPage,omitempty"`
	PageSize     int `json:"pageSize,omitempty"`
	FirstPage    int `json:"firstPage,omitempty"`
	LastPage     int `json:"lastPage,omitempty"`
	TotalRecords int `json:"totalRecords,omitempty"`
}

// Returns a Metadata struct containing metadata for pagination.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
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
