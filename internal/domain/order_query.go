package domain

import "time"

type OrderQueryParams struct {
	Page       int
	Limit      int
	SortBy     string
	SortOrder  string
	Status     string
	CustomerID int64
	StartDate  *time.Time
	EndDate    *time.Time
	Fields     []string
}
