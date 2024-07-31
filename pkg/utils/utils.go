package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidCategoryName = errors.New("invalid category name")
	ErrCategoryNameTooLong = errors.New("category name too long")
	ErrDuplicateCategory   = errors.New("category already exists")
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
