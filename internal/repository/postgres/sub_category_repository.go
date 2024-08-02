package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type subCategoryRepository struct {
	db *sql.DB
}

func NewSubCategoryRepository(db *sql.DB) *subCategoryRepository {
	return &subCategoryRepository{db: db}
}

func (r *subCategoryRepository) Create(ctx context.Context, subCategory *domain.SubCategory) error {
	query := `INSERT INTO sub_categories (parent_category_id, name, slug, gender_specific, created_at) 
			  VALUES ($1, $2, $3, $4, $5) 
			  RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		subCategory.ParentCategoryID,
		subCategory.Name,
		subCategory.Slug,
		subCategory.GenderSpecific,
		subCategory.CreatedAt).Scan(&subCategory.ID)

	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return utils.ErrDuplicateSubCategory
		}
		return err
	}

	return nil
}
