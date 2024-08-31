package domain

import "time"

type SubCategory struct {
	ID               int        `json:"id"`
	ParentCategoryID int        `json:"parent_category_id"`
	Name             string     `json:"name"`
	Slug             string     `json:"slug"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	DeletedAt        *time.Time `json:"deleted_at,omitempty"`
	IsDeleted        bool       `json:"is_deleted"`
}
