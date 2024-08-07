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

func (r *productRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	query := `
		SELECT id, name, description, price, stock_quantity, category_id, sub_category_id, image_url, created_at, updated_at, deleted_at
		FROM products
		WHERE deleted_at IS NULL
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error querying products: %v", err)
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
			&p.CategoryID, &p.SubCategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
			&p.DeletedAt,
		)
		if err != nil {
			log.Printf("Error scanning product row: %v", err)
			return nil, err
		}
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, err
	}

	return products, nil
}
