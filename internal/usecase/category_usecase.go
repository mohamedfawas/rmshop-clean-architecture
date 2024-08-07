package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type CategoryUseCase interface {
	CreateCategory(ctx context.Context, category *domain.Category) error
	GetAllCategories(ctx context.Context) ([]*domain.Category, error)
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

	category.UpdatedAt = time.Now()

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

func (u *categoryUseCase) GetAllCategories(ctx context.Context) ([]*domain.Category, error) {
	log.Println("Entering GetAllCategories use case")
	categories, err := u.categoryRepo.GetAll(ctx)
	if err != nil {
		log.Printf("Failed to retrieve categories: %v", err)
		return nil, err
	}
	log.Printf("Retrieved %d categories from repository", len(categories))
	return categories, nil
}

func (u *categoryUseCase) GetActiveCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	category, err := u.categoryRepo.GetActiveByID(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, errors.New("failed to retrieve category")
	}
	return category, nil
}

func (u *categoryUseCase) UpdateCategory(ctx context.Context, category *domain.Category) error {
	// Trim whitespace from category name
	category.Name = strings.TrimSpace(category.Name)
	category.UpdatedAt = time.Now()

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
		log.Printf("Error in repository.Update: %v", err)
		if err == utils.ErrDuplicateCategory {
			return utils.ErrDuplicateCategory
		}
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		return fmt.Errorf("failed to update category: %v", err)
	}

	return nil
}

func (u *categoryUseCase) SoftDeleteCategory(ctx context.Context, id int) error {
	err := u.categoryRepo.SoftDelete(ctx, id)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		log.Printf("Failed to soft delete category: %v", err)
		return errors.New("failed to delete category")
	}
	return nil
}
