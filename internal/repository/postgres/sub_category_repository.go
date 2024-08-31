package postgres

import (
	"context"
	"database/sql"
	"log"
	"time"

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
	// Insert the new sub-category
	query := `INSERT INTO sub_categories (parent_category_id, name, slug, created_at,updated_at, is_deleted) 
              VALUES ($1, $2, $3, $4,$5, $6) 
              RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		subCategory.ParentCategoryID,
		subCategory.Name,
		subCategory.Slug,
		subCategory.CreatedAt,
		subCategory.UpdatedAt,
		false).Scan(&subCategory.ID)

	if err != nil {
		log.Printf("Error inserting subcategory: %v", err)
		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code == "23505" { // Unique violation error code
				return utils.ErrDuplicateSubCategory
			}
		}
		log.Printf("db error : failed to create sub category : %v", err)
		return err
	}
	return nil
}

func (r *subCategoryRepository) GetByCategoryID(ctx context.Context, categoryID int) ([]*domain.SubCategory, error) {
	query := `
		SELECT id, parent_category_id, name, slug, created_at, deleted_at, is_deleted
		FROM sub_categories
		WHERE parent_category_id = $1 AND is_deleted = FALSE
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query, categoryID)
	if err != nil {
		log.Printf("Error querying sub-categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var subCategories []*domain.SubCategory
	for rows.Next() {
		var sc domain.SubCategory
		err := rows.Scan(
			&sc.ID,
			&sc.ParentCategoryID,
			&sc.Name,
			&sc.Slug,
			&sc.CreatedAt,
			&sc.DeletedAt,
			&sc.IsDeleted,
		)
		if err != nil {
			log.Printf("Error scanning sub-category row: %v", err)
			return nil, err
		}
		subCategories = append(subCategories, &sc)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, err
	}

	return subCategories, nil
}

func (r *subCategoryRepository) GetByID(ctx context.Context, id int) (*domain.SubCategory, error) {
	query := `
		SELECT id, parent_category_id, name, slug, created_at, updated_at, deleted_at, is_deleted
		FROM sub_categories
		WHERE id = $1 AND is_deleted = FALSE
	`

	var subCategory domain.SubCategory
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&subCategory.ID,
		&subCategory.ParentCategoryID,
		&subCategory.Name,
		&subCategory.Slug,
		&subCategory.CreatedAt,
		&subCategory.UpdatedAt,
		&subCategory.DeletedAt,
		&subCategory.IsDeleted,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrSubCategoryNotFound
		}
		log.Printf("Failed to get SubCategory by id=%d: error=%v", id, err)
		return nil, err
	}
	return &subCategory, nil
}

func (r *subCategoryRepository) Update(ctx context.Context, subCategory *domain.SubCategory) error {
	query := `
		UPDATE sub_categories
		SET name = $1, slug = $2, updated_at = $3
		WHERE id = $4 AND parent_category_id = $5 AND is_deleted = FALSE
		RETURNING updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		subCategory.Name,
		subCategory.Slug,
		time.Now().UTC(),
		subCategory.ID,
		subCategory.ParentCategoryID).Scan(&subCategory.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrSubCategoryNotFound
		}
		log.Printf("Error updating sub-category: %v", err)
		return err
	}

	return nil
}

func (r *subCategoryRepository) SoftDelete(ctx context.Context, id int) error {
	query := `
		UPDATE sub_categories
		SET deleted_at = NOW(), is_deleted = TRUE
		WHERE id = $1 AND is_deleted = FALSE
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Error soft deleting sub-category: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrSubCategoryNotFound
	}

	return nil
}
