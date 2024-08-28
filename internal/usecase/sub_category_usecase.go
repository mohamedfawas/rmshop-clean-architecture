package usecase

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/repository"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/validator"
)

type SubCategoryUseCase interface {
	CreateSubCategory(ctx context.Context, categoryID int, subCategory *domain.SubCategory) error
	GetSubCategoriesByCategory(ctx context.Context, categoryID int) ([]*domain.SubCategory, error)
	GetSubCategoryByID(ctx context.Context, categoryID, subCategoryID int) (*domain.SubCategory, error)
	UpdateSubCategory(ctx context.Context, categoryID int, subCategory *domain.SubCategory) error
	SoftDeleteSubCategory(ctx context.Context, categoryID, subCategoryID int) error
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
	parentCategory, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrCategoryNotFound
		}
		log.Printf("database error: failed to retrieve category details for ID %d: %v", categoryID, err)
		return err
	}

	// Generate slug
	subCategory.Slug = utils.GenerateSubCategorySlug(parentCategory.Slug, subCategory.Name)

	// Set creation time and parent category ID
	subCategory.CreatedAt = time.Now().UTC()
	subCategory.UpdatedAt = time.Now().UTC()
	subCategory.ParentCategoryID = categoryID

	// Attempt to create the subcategory
	err = u.subCategoryRepo.Create(ctx, subCategory)
	if err != nil {
		if err == utils.ErrDuplicateSubCategory {
			return utils.ErrDuplicateSubCategory
		}
		return err
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

func (u *subCategoryUseCase) GetSubCategoryByID(ctx context.Context, categoryID, subCategoryID int) (*domain.SubCategory, error) {
	// Check if the parent category exists
	_, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, err
	}

	// Retrieve the sub-category
	subCategory, err := u.subCategoryRepo.GetByID(ctx, subCategoryID)
	if err != nil {
		if err == utils.ErrSubCategoryNotFound {
			return nil, utils.ErrSubCategoryNotFound
		}
		return nil, err
	}

	// Ensure the sub-category belongs to the specified category
	if subCategory.ParentCategoryID != categoryID {
		return nil, utils.ErrSubCategoryNotFound
	}

	return subCategory, nil
}

func (u *subCategoryUseCase) UpdateSubCategory(ctx context.Context, categoryID int, subCategory *domain.SubCategory) error {
	// Check if the parent category exists
	_, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		return err
	}

	// Trim whitespace from subcategory name
	subCategory.Name = strings.TrimSpace(subCategory.Name)

	err = validator.ValidateSubCategoryName(subCategory.Name)
	if err != nil {
		return err
	}

	// Generate slug
	subCategory.Slug = utils.GenerateSlug(subCategory.Name)

	// Set parent category ID
	subCategory.ParentCategoryID = categoryID

	// Attempt to update the subcategory
	err = u.subCategoryRepo.Update(ctx, subCategory)
	if err != nil {
		if err == utils.ErrSubCategoryNotFound {
			return utils.ErrSubCategoryNotFound
		}
		return err
	}

	return nil
}

func (u *subCategoryUseCase) SoftDeleteSubCategory(ctx context.Context, categoryID, subCategoryID int) error {
	// Check if the parent category exists
	_, err := u.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		if err == utils.ErrCategoryNotFound {
			return utils.ErrCategoryNotFound
		}
		return err
	}

	// Check if the sub-category exists and belongs to the specified category
	subCategory, err := u.subCategoryRepo.GetByID(ctx, subCategoryID)
	if err != nil {
		if err == utils.ErrSubCategoryNotFound {
			return utils.ErrSubCategoryNotFound
		}
		return err
	}

	if subCategory.ParentCategoryID != categoryID {
		return utils.ErrSubCategoryNotFound
	}

	// Perform the soft delete
	err = u.subCategoryRepo.SoftDelete(ctx, subCategoryID)
	if err != nil {
		return err
	}

	return nil
}
