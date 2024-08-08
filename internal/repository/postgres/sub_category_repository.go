package postgres

import (
	"context"
	"database/sql"
	"log"

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
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reset the sequence
	_, err = tx.ExecContext(ctx, "SELECT reset_sub_categories_id_seq()")
	if err != nil {
		log.Printf("Error resetting sequence: %v", err)
		return err
	}

	// Insert the new sub-category
	query := `INSERT INTO sub_categories (parent_category_id, name, slug, created_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		subCategory.ParentCategoryID,
		subCategory.Name,
		subCategory.Slug,
		subCategory.CreatedAt).Scan(&subCategory.ID)

	if err != nil {
		log.Printf("Error inserting subcategory: %v", err)
		pqErr, ok := err.(*pq.Error)
		if ok {
			log.Printf("PostgreSQL error code: %s", pqErr.Code)
			if pqErr.Code == "23505" { // Unique violation error code
				return utils.ErrDuplicateSubCategory
			}
		}
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (r *subCategoryRepository) GetByCategoryID(ctx context.Context, categoryID int) ([]*domain.SubCategory, error) {
	query := `
		SELECT id, parent_category_id, name, slug, created_at, deleted_at
		FROM sub_categories
		WHERE parent_category_id = $1 AND deleted_at IS NULL
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
