package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
}

type categoryUseCase struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryUseCase(categoryRepo repository.CategoryRepository) CategoryUseCase {
	return &categoryUseCase{categoryRepo: categoryRepo}
}

func (u *categoryUseCase) CreateCategory(ctx context.Context, category *domain.Category) error {
	// Trim whitespace from category name
	category.Name = strings.TrimSpace(category.Name)

	// Validate category name
	if category.Name == "" {
		return utils.ErrInvalidCategoryName
	}
	if len(category.Name) > 50 {
		return utils.ErrCategoryNameTooLong
	}

	// Generate slug
	category.Slug = utils.GenerateSlug(category.Name)

	// Set creation time
	category.CreatedAt = time.Now()

	// Attempt to create the category
	err := u.categoryRepo.Create(ctx, category)
	if err != nil {
		// Check if it's a duplicate category error (you'd need to implement this check in your repository)
		if err == utils.ErrDuplicateCategory {
			return utils.ErrDuplicateCategory
		}
		// For any other error, return a generic error
		return errors.New("failed to create category")
	}

	return nil
}
