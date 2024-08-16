package utils

import (
	"errors"
	"regexp"
	"strings"
)

var (
	//user related : username
	ErrInvalidUserName         = errors.New("invalid user name")
	ErrUserNameTooShort        = errors.New("user name too short")
	ErrUserNameTooLong         = errors.New("user name too long")
	ErrUserNameWithNumericVals = errors.New("user name with numeric characters")

	//user related : email
	ErrInvalidEmail = errors.New("invalid email format")

	//user related : password
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordTooLong  = errors.New("password too long")
	ErrPasswordInvalid  = errors.New("invalid password")  //empty input string
	ErrPasswordSecurity = errors.New("password not safe") //follow password combination - secure

	//user related : dob
	ErrDOBFormat = errors.New("invalid dob format")

	//user related : dob
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	//category related
	ErrInvalidCategoryName    = errors.New("invalid category name")
	ErrCategoryNameTooLong    = errors.New("category name too long")
	ErrDuplicateCategory      = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrInvalidSubCategoryName = errors.New("invalid subcategory name")
	ErrSubCategoryNameTooLong = errors.New("subcategory name too long")
	ErrDuplicateSubCategory   = errors.New("subcategory already exists")
	ErrSubCategoryNotFound    = errors.New("subcategory not found")
	ErrCategoryNameTooShort   = errors.New("category name too short")
	ErrCategoryNameNumeric    = errors.New("category name purely numeric")
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
