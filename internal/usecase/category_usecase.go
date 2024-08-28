package usecase

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context, category *domain.Category) error // fz
	GetAllCategories(ctx context.Context) ([]*domain.Category, error)    // fz
	GetActiveCategoryByID(ctx context.Context, id int) (*domain.Category, error)
	UpdateCategory(ctx context.Context, category *domain.Category) error
	SoftDeleteCategory(ctx context.Context, id int) error
}

type categoryUseCase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUseCase(categoryRepo repository.CategoryRepository) CategoryUseCase {
	return &categoryUseCase{categoryRepo: categoryRepo}
}

// CreateCategory adds a new category to the repository.
//
// This method sets the `Slug`, `CreatedAt`, and `UpdatedAt` fields of the
// provided category before attempting to persist it in the database. It
// handles the case where a category with the same name already exists.
//
// Parameters:
//   - ctx: The context for managing request deadlines, cancellation signals,
//     and other request-scoped values.
//   - category: A pointer to a Category object that contains the details of
//     the category to be created.
//
// Returns:
//   - error: An error if the category could not be created due to reasons
//     such as a duplicate entry; otherwise, nil.
//
// Notes:
//   - The `Slug` is generated using a utility function based on the category's
//     name.
//   - The `CreatedAt` and `UpdatedAt` fields are both set to the current UTC time.
//   - If the category already exists, a specific error (utils.ErrDuplicateCategory)
//     is returned.
func (u *categoryUseCase) CreateCategory(ctx context.Context, category *domain.Category) error {

	// Generate slug
	category.Slug = utils.GenerateSlug(category.Name)

	// Set creation time
	category.CreatedAt = time.Now().UTC()

	// set category updation time
	category.UpdatedAt = time.Now().UTC()

	// Attempt to create the category in the database
	err := u.categoryRepo.Create(ctx, category)
	if err != nil {
		if err == utils.ErrDuplicateCategory {
			return utils.ErrDuplicateCategory
		}
		return err
	}

	return nil
}

// GetAllCategories retrieves all categories from the repository.
//
// This method interacts with the category repository to fetch a list of all
// categories available in the system.
//
// Parameters:
//   - ctx: The context for managing request deadlines, cancelation signals,
//     and other request-scoped values.
//
// Returns:
//   - []*domain.Category: A slice of pointers to Category objects representing
//     the categories retrieved from the repository.
//   - error: An error if the categories could not be retrieved; otherwise nil.
func (u *categoryUseCase) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	categories, err := u.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// GetActiveCategoryByID retrieves a category by its ID from the repository.
//
// This method attempts to fetch a category by its unique identifier. If the
// category is not found, a specific error is returned. It also handles
// other potential errors that may occur during the retrieval process.
//
// Parameters:
//   - ctx: The context for managing request deadlines, cancellation signals,
//     and other request-scoped values.
//   - id: The unique identifier of the category to retrieve.
//
// Returns:
//   - *domain.Category: A pointer to the Category object representing the
//     retrieved category, or nil if not found.
//   - error: An error if the category could not be retrieved due to reasons
//     such as not being found or other repository-related errors.
//
// Notes:
//   - If the category is not found in the repository, the method returns a
//     specific error (utils.ErrCategoryNotFound).
func (u *categoryUseCase) GetActiveCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	category, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, err
	}
	return category, nil
}

// UpdateCategory modifies an existing category in the repository.
//
// This method updates the provided category by trimming whitespace from its name,
// validating the name length, generating a new slug, and setting the `UpdatedAt`
// field to the current time. It then attempts to update the category in the
// repository and handles specific errors such as duplicate categories or
// non-existent categories.
//
// Parameters:
//   - ctx: The context for managing request deadlines, cancellation signals,
//     and other request-scoped values.
//   - category: A pointer to a Category object containing the updated details
//     of the category.
//
// Returns:
//   - error: An error if the category could not be updated due to reasons
//     such as invalid input, duplicate entries, or the category not
//     being found; otherwise, nil.
//
// Notes:
//   - The category name is trimmed of any leading or trailing whitespace.
//   - If the category name is empty or exceeds 50 characters, specific errors
//     are returned and logged.
//   - The `Slug` is regenerated based on the updated category name.
//   - If the category already exists, a specific error (utils.ErrDuplicateCategory)
//     is returned. If the category to update is not found, another specific error
//     (utils.ErrCategoryNotFound) is returned.
func (u *categoryUseCase) UpdateCategory(ctx context.Context, category *domain.Category) error {
	// Trim whitespace from category name
	category.Name = strings.TrimSpace(category.Name)
	category.UpdatedAt = time.Now().UTC()

	// Validate category name
	if category.Name == "" {
		log.Printf("Invalid category name: empty name")
		return utils.ErrInvalidCategoryName
	}
	if len(category.Name) > 50 {
		log.Printf("Category name too long: %d characters", len(category.Name))
		return utils.ErrCategoryNameTooLong
	}

	// Generate slug
	category.Slug = utils.GenerateSlug(category.Name)

	// Attempt to update the category
	err := u.categoryRepo.Update(ctx, category)
	if err != nil {
		if err == utils.ErrDuplicateCategory {
			return utils.ErrDuplicateCategory
		}
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		return err
	}

	return nil
}

// SoftDeleteCategory marks a category as deleted without removing it from the repository.
//
// This method performs a soft delete by updating the category's status or a similar
// attribute to indicate that it has been deleted, while preserving the record in
// the database. It handles specific errors, including cases where the category
// does not exist.
//
// Parameters:
//   - ctx: The context for managing request deadlines, cancellation signals,
//     and other request-scoped values.
//   - id: The unique identifier of the category to be soft deleted.
//
// Returns:
//   - error: An error if the category could not be soft deleted due to reasons
//     such as the category not being found; otherwise, nil.
//
// Notes:
//   - If the category is not found in the repository, a specific error
//     (utils.ErrCategoryNotFound) is returned.
//   - Any other errors encountered during the soft delete operation are logged
//     and returned.
func (u *categoryUseCase) SoftDeleteCategory(ctx context.Context, id int) error {
	err := u.categoryRepo.SoftDelete(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		log.Printf("Failed to soft delete category: %v", err)
		return err
	}
	return nil
}
