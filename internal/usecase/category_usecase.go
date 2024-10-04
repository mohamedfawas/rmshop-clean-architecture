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

func (u *categoryUseCase) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	categories, err := u.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

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

func (u *categoryUseCase) SoftDeleteCategory(ctx context.Context, id int) error {
	category, err := u.categoryRepo.GetByID(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		log.Printf("Failed to retrieve category: %v", err)
		return err
	}

	if category.IsDeleted {
		return utils.ErrCategoryAlreadyDeleted
	}

	err = u.categoryRepo.SoftDelete(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		log.Printf("Failed to soft delete category: %v", err)
		return err
	}
	return nil
}
