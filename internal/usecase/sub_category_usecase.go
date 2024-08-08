package usecase

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type SubCategoryUseCase interface {
	CreateSubCategory(ctx context.Context, categoryID int, subCategory *domain.SubCategory) error
	GetSubCategoriesByCategory(ctx context.Context, categoryID int) ([]*domain.SubCategory, error)
}

type subCategoryUseCase struct {
	subCategoryRepo repository.SubCategoryRepository
	categoryRepo    repository.CategoryRepository
}

func NewSubCategoryUseCase(subCategoryRepo repository.SubCategoryRepository, categoryRepo repository.CategoryRepository) SubCategoryUseCase {
	return &subCategoryUseCase{
		subCategoryRepo: subCategoryRepo,
		categoryRepo:    categoryRepo,
	}
}

func (u *subCategoryUseCase) CreateSubCategory(ctx context.Context, categoryID int, subCategory *domain.SubCategory) error {
	// Check if the parent category exists
	_, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrCategoryNotFound
		}
		return err
	}

	// Trim whitespace from subcategory name
	subCategory.Name = strings.TrimSpace(subCategory.Name)

	// Validate subcategory name
	if subCategory.Name == "" {
		return utils.ErrInvalidSubCategoryName
	}
	if len(subCategory.Name) > 50 {
		return utils.ErrSubCategoryNameTooLong
	}

	// Generate slug
	subCategory.Slug = utils.GenerateSlug(subCategory.Name)

	// Set creation time and parent category ID
	subCategory.CreatedAt = time.Now()
	subCategory.ParentCategoryID = categoryID

	// Attempt to create the subcategory
	err = u.subCategoryRepo.Create(ctx, subCategory)
	if err != nil {
		if err == utils.ErrDuplicateSubCategory {
			return utils.ErrDuplicateSubCategory
		}
		return errors.New("failed to create subcategory")
	}

	return nil
}

func (u *subCategoryUseCase) GetSubCategoriesByCategory(ctx context.Context, categoryID int) ([]*domain.SubCategory, error) {
	// Check if the parent category exists
	_, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, errors.New("failed to retrieve parent category")
	}

	// Retrieve sub-categories
	subCategories, err := u.subCategoryRepo.GetByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, errors.New("failed to retrieve sub-categories")
	}

	return subCategories, nil
}
