package domain

import "time"

type Category struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"` // * used to represent it as either a time stamp or null value
	IsDeleted bool       `json:"is_deleted"`
}
