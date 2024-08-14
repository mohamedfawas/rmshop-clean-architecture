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

	query := `INSERT INTO products (name, description, price, stock_quantity, category_id, sub_category_id, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id`

	err = tx.QueryRowContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity,
		product.CategoryID, product.SubCategoryID, product.CreatedAt, product.UpdatedAt).Scan(&product.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
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
        SELECT p.id, p.name, p.description, p.price, p.stock_quantity, 
               p.category_id, p.sub_category_id, p.created_at, p.updated_at, 
               p.deleted_at, p.primary_image_id
        FROM products p
        WHERE p.id = $1 AND p.deleted_at IS NULL
    `

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
		&p.CategoryID, &p.SubCategoryID, &p.CreatedAt, &p.UpdatedAt,
		&p.DeletedAt, &p.PrimaryImageID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		log.Printf("Error querying product: %v", err)
		return nil, err
	}

	// Fetch product images
	images, err := r.GetProductImages(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Images = images

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

func (r *productRepository) GetActiveProducts(ctx context.Context, page, pageSize int) ([]*domain.Product, int, error) {
	query := `
        SELECT id, name, description, price, stock_quantity, category_id, sub_category_id, image_url, created_at, updated_at
        FROM products
        WHERE deleted_at IS NULL
        ORDER BY id
        LIMIT $1 OFFSET $2
    `

	countQuery := `
        SELECT COUNT(*)
        FROM products
        WHERE deleted_at IS NULL
    `

	offset := (page - 1) * pageSize

	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		log.Printf("Error querying active products: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
			&p.CategoryID, &p.SubCategoryID, &p.ImageURL, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning product row: %v", err)
			return nil, 0, err
		}
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error after scanning all rows: %v", err)
		return nil, 0, err
	}

	var totalCount int
	err = r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		log.Printf("Error getting total count: %v", err)
		return nil, 0, err
	}

	return products, totalCount, nil
}

func (r *productRepository) AddProductImages(ctx context.Context, productID int64, images []domain.ProductImage) ([]int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var imageIDs []int64
	for _, img := range images {
		var imageID int64
		err := tx.QueryRowContext(ctx,
			"INSERT INTO product_images (product_id, image_url) VALUES ($1, $2) RETURNING id",
			productID, img.ImageURL).Scan(&imageID)
		if err != nil {
			return nil, err
		}
		imageIDs = append(imageIDs, imageID)
	}

	return imageIDs, tx.Commit()
}

func (r *productRepository) GetProductImages(ctx context.Context, productID int64) ([]domain.ProductImage, error) {
	query := `
        SELECT id, product_id, image_url, is_primary, created_at 
        FROM product_images 
        WHERE product_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.ProductImage
	for rows.Next() {
		var img domain.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ImageURL, &img.IsPrimary, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *productRepository) UpdatePrimaryImage(ctx context.Context, productID int64, imageID int64) error {
	query := `
        UPDATE products
        SET primary_image_id = $1
        WHERE id = $2 AND deleted_at IS NULL
    `

	result, err := r.db.ExecContext(ctx, query, imageID, productID)
	if err != nil {
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

func (r *productRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

func (r *productRepository) CreateWithTx(ctx context.Context, tx *sql.Tx, product *domain.Product) error {
	query := `INSERT INTO products (name, description, price, stock_quantity, category_id, sub_category_id, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
              RETURNING id`

	err := tx.QueryRowContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity,
		product.CategoryID, product.SubCategoryID, product.CreatedAt, product.UpdatedAt).Scan(&product.ID)
	return err
}

func (r *productRepository) AddProductImagesWithTx(ctx context.Context, tx *sql.Tx, productID int64, images []domain.ProductImage) ([]int64, error) {
	var imageIDs []int64
	for _, img := range images {
		var imageID int64
		err := tx.QueryRowContext(ctx,
			"INSERT INTO product_images (product_id, image_url, is_primary) VALUES ($1, $2, $3) RETURNING id",
			productID, img.ImageURL, img.IsPrimary).Scan(&imageID)
		if err != nil {
			return nil, err
		}
		imageIDs = append(imageIDs, imageID)
	}
	return imageIDs, nil
}

func (r *productRepository) UpdatePrimaryImageWithTx(ctx context.Context, tx *sql.Tx, productID int64, imageID int64) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE products SET primary_image_id = $1 WHERE id = $2",
		imageID, productID)
	return err
}
