package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

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

func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `
		SELECT id, name, description, price, stock_quantity, category_id, sub_category_id, image_url, created_at, updated_at, deleted_at
		FROM products
		WHERE id = $1 AND deleted_at IS NULL
	`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
		&p.CategoryID, &p.SubCategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
		&p.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		log.Printf("Error querying product: %v", err)
		return nil, err
	}

	return &p, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock_quantity = $4,
			category_id = $5, sub_category_id = $6, image_url = $7, updated_at = $8
		WHERE id = $9 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity,
		product.CategoryID, product.SubCategoryID, product.ImageURL,
		time.Now(), product.ID)
	if err != nil {
		log.Printf("Error updating product: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found or already deleted")
	}

	return nil
}

func (r *productRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `
		UPDATE products
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		log.Printf("Error soft deleting product: %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("product not found or already deleted")
	}

	return nil
}
