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

type categoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) *categoryRepository {
	return &categoryRepository{db: db}
}

// Create inserts a new category into the database
// It accepts a context for request scoping and a Category struct containing the category details
// If a category with the same name or slug already exists, it returns a duplicate category error
func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {

	// SQL query to insert a new category and return the new category's ID
	query := `INSERT INTO categories (name, slug, created_at,updated_at) 
              VALUES ($1, $2, $3, $4) 
              RETURNING id`

	// Execute the query with the provided category details and scan the returned ID into the category struct
	err := r.db.QueryRowContext(ctx, query,
		category.Name, category.Slug, category.CreatedAt, category.UpdatedAt).Scan(&category.ID)
	if err != nil {
		// Type assert the error to a PostgreSQL error to handle specific database errors
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Check for unique constraint violation error (duplicate entry)
			return utils.ErrDuplicateCategory
		}
		log.Printf("error while creating category : %v", err)
		return err
	}
	return nil
}

// GetByID retrieves a category from the database by its ID
// It returns a pointer to the Category struct if found, or an error if the category does not exist, soft deleted or an issue occurs during the query
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
		log.Printf("error while retrieving category using id : %v", err)
		return nil, err
	}

	return &category, nil
}

// GetAll retrieves all categories from the database that are not soft-deleted (deleted_at IS NULL)
// It returns a slice of pointers to Category structs, or an error if the query fails or issues occur during row iteration
func (r *categoryRepository) GetAll(ctx context.Context) ([]*domain.Category, error) {
	query := `SELECT id, name, slug, created_at, deleted_at,updated_at FROM categories WHERE deleted_at IS NULL ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("error while retrieving category details : %v", err)
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
			&category.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning row into category struct: %v", err)
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error encountered during rows iteration: %v", err)
		return nil, err
	}
	return categories, nil
}

// Update modifies an existing category in the database based on the provided Category struct
// It returns an error if the update fails, the category is not found, or a unique constraint is violated
func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	query := `UPDATE categories SET name = $1, slug = $2, updated_at = NOW() WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, category.Name, category.Slug, category.ID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code == "23505" { // Unique violation error code
				return utils.ErrDuplicateCategory
			}
		}
		log.Printf("Error executing query to retrieve category details: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No rows affected, category not found")
		return utils.ErrCategoryNotFound
	}

	return nil
}

// SoftDelete marks a category as deleted by setting the deleted_at timestamp
// It returns an error if the operation fails or the category does not exist or is already deleted
func (r *categoryRepository) SoftDelete(ctx context.Context, id int) error {
	query := `UPDATE categories SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		log.Printf("Error executing soft delete query: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
		return err
	}

	if rowsAffected == 0 {
		log.Printf("No rows affected, category not found or already deleted")
		return utils.ErrCategoryNotFound
	}

	return nil
}
