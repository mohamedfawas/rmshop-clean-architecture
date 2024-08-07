package postgres

import (
	"context"
	"database/sql"
	"fmt"
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

func (r *categoryRepository) GetAll(ctx context.Context) ([]*domain.Category, error) {
	log.Println("Entering GetAll repository method")
	query := `SELECT id, name, slug, created_at, deleted_at FROM categories WHERE deleted_at IS NULL ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var categories []*domain.Category
	for rows.Next() {
		var category domain.Category
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.CreatedAt,
			&category.DeletedAt,
		)
		if err != nil {
			log.Printf("Error scanning category row: %v", err)
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d categories from database", len(categories))
	return categories, nil
}

func (r *categoryRepository) GetActiveByID(ctx context.Context, id int) (*domain.Category, error) {
	query := `SELECT id, name, slug, created_at, deleted_at FROM categories WHERE id = $1 AND deleted_at IS NULL`

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

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `UPDATE categories SET name = $1, slug = $2, updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL`

	log.Printf("Executing update query: %s", query)
	log.Printf("Query parameters: name=%s, slug=%s, id=%d", category.Name, category.Slug, category.ID)

	result, err := r.db.ExecContext(ctx, query, category.Name, category.Slug, category.ID)
	if err != nil {
		log.Printf("Error executing update query: %v", err)
		pqErr, ok := err.(*pq.Error)
		if ok {
			log.Printf("PostgreSQL error code: %s", pqErr.Code)
			if pqErr.Code == "23505" { // Unique violation error code
				return utils.ErrDuplicateCategory
			}
		}
		return fmt.Errorf("database error: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	log.Printf("Rows affected: %d", rowsAffected)

	if rowsAffected == 0 {
		log.Printf("No rows affected, category not found")
		return utils.ErrCategoryNotFound
	}

	return nil
}
