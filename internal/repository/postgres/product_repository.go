package postgres

import (
	"context"
	"database/sql"
	"log"

	"github.com/mohamedfawas/rmshop-clean-architecture/internal/domain"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO products (name, description, price, stock_quantity, category_id, sub_category_id, image_url, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity,
		product.CategoryID, product.SubCategoryID, product.ImageURL,
		product.CreatedAt, product.UpdatedAt).Scan(&product.ID)
	if err != nil {
		log.Printf("Error creating product: %v", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}
