package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *categoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reset the sequence
	_, err = tx.ExecContext(ctx, "SELECT reset_categories_id_seq()")
	if err != nil {
		log.Printf("Error resetting sequence: %v", err)
		return err
	}

	query := `INSERT INTO categories (name, slug, created_at) 
              VALUES ($1, $2, $3) 
              RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		category.Name, category.Slug, category.CreatedAt).Scan(&category.ID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return utils.ErrDuplicateCategory
		}
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

// Add the GetByID method
func (r *categoryRepository) GetByID(ctx context.Context, id int) (*domain.Category, error) {
	query := `SELECT id, name, slug, created_at, deleted_at FROM categories WHERE id = $1`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.CreatedAt,
		&category.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrCategoryNotFound
		}
		return nil, err
	}

	return &category, nil
}
