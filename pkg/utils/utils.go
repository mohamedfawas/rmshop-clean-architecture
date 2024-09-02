package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/lib/pq"
)

// usecase constants
const (
	MaxResendAttempts   = 3
	ResendCooldown      = 1 * time.Minute
	SignupExpiration    = 1 * time.Hour
	MaxImagesPerProduct = 5
	MaxFileSize         = 10 * 1024 * 1024 // 10 MB
	MaxCartItemQuantity = 10
	MaxDiscountAmount   = 5000
)

const (
	InternalServerErrorString = "Internal server error"
)

var (
	//user : name
	ErrInvalidUserName         = errors.New("invalid user name")
	ErrUserNameTooShort        = errors.New("user name too short")
	ErrUserNameTooLong         = errors.New("user name too long")
	ErrUserNameWithNumericVals = errors.New("user name with numeric characters")

	//user : email
	ErrInvalidEmail = errors.New("invalid email format")
	ErrMissingEmail = errors.New("no email input given")

	//user : password
	ErrPasswordTooShort = errors.New("password too short")
	ErrPasswordTooLong  = errors.New("password too long")
	ErrPasswordInvalid  = errors.New("invalid password")  //empty input string
	ErrPasswordSecurity = errors.New("password not safe") //follow password combination - secure

	//user : dob
	ErrDOBFormat = errors.New("invalid dob format")

	//user : phone number
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	//user : OTP
	ErrMissingOTP                   = errors.New("need valid otp input")
	ErrOtpLength                    = errors.New("otp must have 6 digits")
	ErrOtpNums                      = errors.New("non digits in otp")
	ErrTooManyResendAttempts        = errors.New("too many resend attempts")
	ErrSignupExpired                = errors.New("signup process has expired")
	ErrRetrieveOTPResendInfo        = errors.New("error retrieving otp resend info")
	ErrUpdateVerficationAfterResend = errors.New("error update verification entry after resend otp")
	ErrUpdateOTPResendTable         = errors.New("error updating otp resend table")
	ErrCreateVerificationEntry      = errors.New("error creating verification entry")
	ErrSendingResetToken            = errors.New("error sending reset token")

	//user : db operations
	ErrUserNotFound              = errors.New("user not found")
	ErrOTPNotFound               = errors.New("OTP not found")
	ErrDuplicateEmail            = errors.New("email already exists")
	ErrVerificationEntryNotFound = errors.New("verification entry not found")

	//user : login
	ErrLoginCredentialsMissing  = errors.New("login credentials missing")
	ErrCreateUser               = errors.New("failed to create user")
	ErrUpdateVerificationEntry  = errors.New("error updating verification entry")
	ErrDeleteVerificationEntry  = errors.New("error deleting verification entry")
	ErrUpdateLastLogin          = errors.New("error updating last login")
	ErrGenerateJWTTokenWithRole = errors.New("error generate jwt token with role")
	ErrVerificationEntryType    = errors.New("wrong verification entry type")

	//user : logout
	ErrMissingAuthToken         = errors.New("missing auth token")
	ErrAuthHeaderFormat         = errors.New("invalid auth header format")
	ErrEmptyToken               = errors.New("empty token")
	ErrFailedToCheckBlacklisted = errors.New("failed to check if token is blacklisted")

	// user : update profile
	ErrNoUpdateData   = errors.New("no update data found")
	ErrUpdateUserData = errors.New("error updating user data")

	//admin : login
	ErrMissingAdminCredentials = errors.New("admin username and password missing")
	ErrAdminUsernameTooLong    = errors.New("admin username too long")
	ErrAdminPasswordTooLong    = errors.New("admin password too long")
	ErrRetreivingAdminUsername = errors.New("error retrieving admin username")
	ErrCheckTokenBlacklisted   = errors.New("failed to check if token is blacklisted")
	ErrInvalidExpirationClaim  = errors.New("invalid expiration claim")
	ErrTokenExpired            = errors.New("token expired")

	// user address
	ErrUserAddressTooLong       = errors.New("user address too long")
	ErrUserAddressTooShort      = errors.New("user address too short")
	ErrInvalidUserStateEntry    = errors.New("invalid state")
	ErrInvalidUserCityEntry     = errors.New("invalid city")
	ErrInvalidPinCode           = errors.New("invalid pin code")
	ErrUserAddressAlreadyExists = errors.New("user address already exists")
	ErrAddressNotFound          = errors.New("address not found")
	ErrUnauthorized             = errors.New("unauthorized to updated this address")

	//category
	ErrInvalidCategoryName    = errors.New("invalid category name")
	ErrCategoryNameTooLong    = errors.New("category name too long")
	ErrDuplicateCategory      = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryNameTooShort   = errors.New("category name too short")
	ErrCategoryNameNumeric    = errors.New("category name purely numeric")
	ErrDBCreateCategory       = errors.New("failed to create category in db")
	ErrCategoryAlreadyDeleted = errors.New("category already deleted")

	//sub category
	ErrInvalidSubCategoryName    = errors.New("invalid subcategory name")
	ErrSubCategoryNameTooLong    = errors.New("subcategory name too long")
	ErrSubCategoryNotFound       = errors.New("subcategory not found")
	ErrDuplicateSubCategory      = errors.New("subcategory already exists")
	ErrSubCategoryNameTooShort   = errors.New("sub category name too short")
	ErrSubCategoryNameNumeric    = errors.New("sub category name purely numeric")
	ErrCreateSubCategory         = errors.New("failed to create sub category")
	ErrSubCategoryAlreadyDeleted = errors.New("sub category already deleted")

	//db errors
	ErrQueryExecution    = errors.New("failed to execute query")
	ErrRowScan           = errors.New("failed to scan row")
	ErrNoCategoriesFound = errors.New("no category found")

	// product related
	ErrInvalidProductName         = errors.New("invalid product name")
	ErrProductNameTooLong         = errors.New("product name is too long")
	ErrProductNameTooShort        = errors.New("product name too short")
	ErrInvalidProductDescription  = errors.New("invalid product description")
	ErrInvalidProductPrice        = errors.New("invalid product price")
	ErrInvalidStockQuantity       = errors.New("invalid stock quantity")
	ErrInvalidCategoryID          = errors.New("invalid category ID")
	ErrInvalidSubCategoryID       = errors.New("invalid sub-category ID")
	ErrNoImages                   = errors.New("at least one image is required")
	ErrInvalidImageURL            = errors.New("invalid image URL")
	ErrMultiplePrimaryImages      = errors.New("multiple primary images not allowed")
	ErrNoPrimaryImage             = errors.New("no primary image specified")
	ErrStockQuantRequired         = errors.New("stock quantity is required")
	ErrProductDescriptionRequired = errors.New("product description required")
	ErrDuplicateProductName       = errors.New("product name already exists")
	ErrDuplicateProductSlug       = errors.New("product slug already exists")
	ErrInvalidQueryParameter      = errors.New("invalid query parameter")

	//usecase errors
	ErrAdminNotFound           = errors.New("admin not found")
	ErrInvalidAdminCredentials = errors.New("invalid admin credentials")
	ErrInvalidAdminToken       = errors.New("invalid admin token")
	//ErrDuplicateEmail          = errors.New("email already exists")
	//ErrUserNotFound            = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	//ErrOTPNotFound             = errors.New("OTP not found")
	ErrInvalidOTP           = errors.New("invalid OTP")
	ErrExpiredOTP           = errors.New("OTP has expired")
	ErrEmailAlreadyVerified = errors.New("email already verified")
	ErrInvalidInput         = errors.New("invalid input")
	ErrDatabaseUnavailable  = errors.New("database unavailable")
	ErrSMTPServerIssue      = errors.New("SMTP server issue")
	ErrNonExEmail           = errors.New("OTP not found for given email") //non existent email
	//ErrSignupExpired           = errors.New("signup process has expired")
	//ErrTooManyResendAttempts   = errors.New("too many resend attempts")
	ErrUserBlocked             = errors.New("user is blocked")
	ErrTokenAlreadyBlacklisted = errors.New("token already blacklisted")
	ErrInvalidCategory         = errors.New("invalid category ID")
	ErrInvalidSubCategory      = errors.New("invalid sub-category ID")
	ErrProductNotFound         = errors.New("product not found")
	ErrDuplicateImageURL       = errors.New("duplicate image URL")
	ErrCreateVericationEntry   = errors.New("error creating verification entry")
	ErrSendingOTP              = errors.New("error sending OTP email")
	ErrGenerateOTP             = errors.New("error generating otp")
	ErrHashingPassword         = errors.New("error hashing password")

	//image
	ErrImageNotFound   = errors.New("image not found")
	ErrLastImage       = errors.New("last image")
	ErrFileTooLarge    = errors.New("file size exceeds the maximum limit of 10MB")
	ErrInvalidFileType = errors.New("invalid file type. Only .jpg, .jpeg, .png, and .gif are allowed")
	ErrTooManyImages   = errors.New("maximum number of images (5) reached for this product")
	ErrEmptyFile       = errors.New("file is empty")

	// user password change
	ErrOTPNotRequested      = errors.New("no OTP was requested for this email")
	ErrTooManyResetAttempts = errors.New("too many reset attempts")
	ErrSamePassword         = errors.New("new password cannot be the same as the old password")

	// user address
	ErrLastAddress       = errors.New("last address")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("invalid quantity")
	ErrCartFull          = errors.New("cart if full")

	// cart
	ErrCartItemNotFound   = errors.New("cart item not found")
	ErrExceedsMaxQuantity = errors.New("exceeds maximum quantity")
	ErrEmptyCart          = errors.New("empty cart")

	// token errors
	ErrUnexpectedSigning = errors.New("unexpected signing method")
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidRole       = errors.New("invalid role")

	// coupon
	ErrCouponNotFound            = errors.New("coupon not found")
	ErrCouponInactive            = errors.New("coupon inactive")
	ErrCouponExpired             = errors.New("coupon expired")
	ErrOrderTotalBelowMinimum    = errors.New("order total below minimum")
	ErrDuplicateCouponCode       = errors.New("duplicate coupon code")
	ErrInvalidCouponCode         = errors.New("invalid coupon code")
	ErrInvalidDiscountPercentage = errors.New("invalid discount percentage")
	ErrInvalidMinOrderAmount     = errors.New("invalid minimum order amount")
	ErrInvalidExpiryDate         = errors.New("invalid expiry date")
	ErrCouponAlreadyApplied      = errors.New("coupon already applied")

	ErrInternalServer = errors.New("internal server error")
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

func GenerateSubCategorySlug(categorySlug, subCategoryName string) string {
	// Convert sub-category name to lowercase
	subCategorySlug := strings.ToLower(subCategoryName)

	// Replace spaces with hyphens
	subCategorySlug = strings.ReplaceAll(subCategorySlug, " ", "-")

	// Remove any characters that are not alphanumeric or hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	subCategorySlug = reg.ReplaceAllString(subCategorySlug, "")

	// Remove leading and trailing hyphens
	subCategorySlug = strings.Trim(subCategorySlug, "-")

	// Combine category slug with sub-category slug
	return fmt.Sprintf("%s/%s", categorySlug, subCategorySlug)
}

// IsDuplicateKeyError checks if the given error is a database error
// indicating a duplicate key violation (usually due to a unique constraint).
func IsDuplicateKeyError(err error) bool {
	// Implementation for lib/pq
	if pqErr, ok := err.(*pq.Error); ok {
		return pqErr.Code == "23505" // 23505 is the PostgreSQL error code for unique_violation
	}

	// Generic check (less reliable, but can work as a fallback)
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
