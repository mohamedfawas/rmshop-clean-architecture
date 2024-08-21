package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/lib/pq"
	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/utils"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (name, slug, description, price, stock_quantity, sub_category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		product.Name, product.Slug, product.Description, product.Price, product.StockQuantity,
		product.SubCategoryID, product.CreatedAt, product.UpdatedAt).Scan(&product.ID)

	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code {
			case "23505": // Unique violation error code
				if pqErr.Constraint == "products_name_key" {
					return utils.ErrDuplicateProductName
				}
				if pqErr.Constraint == "products_slug_key" {
					return utils.ErrDuplicateProductSlug
				}
			}
		}
		log.Printf("database error while creating product: %v", err)
		return err
	}

	return nil
}

func (r *productRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	// this query returns a boolean value
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if slug exists, %v:", err)
	}
	return exists, err
}

func (r *productRepository) NameExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE LOWER(name) = LOWER($1))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if product name exists: %v", err)
	}
	return exists, err
}
