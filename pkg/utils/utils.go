package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidCategoryName    = errors.New("invalid category name")
	ErrCategoryNameTooLong    = errors.New("category name too long")
	ErrDuplicateCategory      = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrInvalidSubCategoryName = errors.New("invalid subcategory name")
	ErrSubCategoryNameTooLong = errors.New("subcategory name too long")
	ErrDuplicateSubCategory   = errors.New("subcategory already exists")
	ErrSubCategoryNotFound    = errors.New("subcategory not found")
)

func GenerateSlug(name string) string {
	// Convert to lowercase
	slug := strings.ToLower(name)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any characters that are not alphanumeric or hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}
