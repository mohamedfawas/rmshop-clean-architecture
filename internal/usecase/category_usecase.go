package usecase

import (
	"context"
	"errors"
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
