package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

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
	query := `INSERT INTO categories (name, slug, created_at,updated_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		category.Name, category.Slug, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation error code
			return utils.ErrDuplicateCategory
		}
		return utils.ErrDBCreateCategory
	}
	return nil
}

// Add the GetByID method
func (r *categoryRepository) GetByID(ctx context.Context, id int) (*domain.Category, error) {
	query := `SELECT id, name, slug, created_at, updated_at, deleted_at FROM categories WHERE id = $1 AND deleted_at IS NULL`

	var category domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.CreatedAt,
		&category.UpdatedAt,
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
	query := `SELECT id, name, slug, created_at, deleted_at,updated_at FROM categories WHERE deleted_at IS NULL ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, utils.ErrQueryExecution
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
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, utils.ErrRowScan
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		return nil, utils.ErrQueryExecution
	}
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

func (r *categoryRepository) SoftDelete(ctx context.Context, id int) error {
	query := `UPDATE categories SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		log.Printf("Error executing soft delete query: %v", err)
		return fmt.Errorf("database error: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return fmt.Errorf("error checking rows affected: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("No rows affected, category not found or already deleted")
		return utils.ErrCategoryNotFound
	}

	return nil
}
