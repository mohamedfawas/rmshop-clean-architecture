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

func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `SELECT id, name, slug, description, price, stock_quantity, sub_category_id, created_at, updated_at, deleted_at
              FROM products WHERE id = $1 AND deleted_at IS NULL`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description, &product.Price,
		&product.StockQuantity, &product.SubCategoryID, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrProductNotFound
		}
		log.Printf("Database error while retrieving product by ID: %v", err)
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `UPDATE products SET name = $1, slug = $2, description = $3, price = $4, 
              stock_quantity = $5, sub_category_id = $6, updated_at = $7
              WHERE id = $8 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, product.Name, product.Slug, product.Description,
		product.Price, product.StockQuantity, product.SubCategoryID, time.Now(), product.ID)

	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code {
			case "23505": // Unique violation
				return utils.ErrDuplicateProductName
			}
		}
		log.Printf("Database error while updating product details, %v: ", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error while checking rows affected %v:", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrProductNotFound
	}

	return nil
}

func (r *productRepository) SoftDelete(ctx context.Context, id int64) error {
	query := `UPDATE products SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrProductNotFound
	}

	return nil
}

func (r *productRepository) NameExistsBeforeUpdate(ctx context.Context, name string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE LOWER(name) = LOWER($1) AND id != $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name, excludeID).Scan(&exists)
	return exists, err
}

func (r *productRepository) AddImage(ctx context.Context, productID int64, imageURL string, isPrimary bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// If this is a primary image, update all other images to non-primary
	if isPrimary {
		_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = false WHERE product_id = $1", productID)
		if err != nil {
			return err
		}
	}
	// Insert the new image
	query := `
		INSERT INTO product_images (product_id, image_url, is_primary)
		VALUES ($1, $2, $3)
	`

	_, err = tx.ExecContext(ctx, query, productID, imageURL, isPrimary)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation
			return utils.ErrDuplicateImageURL
		}
		return err
	}
	// If this is a primary image, update the product's primary_image_id
	if isPrimary {
		_, err = tx.ExecContext(ctx, "UPDATE products SET primary_image_id = (SELECT id FROM product_images WHERE product_id = $1 AND is_primary = true) WHERE id = $1", productID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *productRepository) GetImageCount(ctx context.Context, productID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM product_images WHERE product_id = $1", productID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *productRepository) DeleteImage(ctx context.Context, productID int64, imageURL string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM product_images WHERE product_id = $1 AND image_url = $2", productID, imageURL)
	return err
}
func (r *productRepository) GetImageByURL(ctx context.Context, productID int64, imageURL string) (*domain.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, created_at
		FROM product_images
		WHERE product_id = $1 AND image_url = $2
	`
	var image domain.ProductImage
	err := r.db.QueryRowContext(ctx, query, productID, imageURL).Scan(
		&image.ID, &image.ProductID, &image.ImageURL, &image.IsPrimary, &image.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrImageNotFound
		}
		return nil, err
	}
	return &image, nil
}

func (r *productRepository) GetProductImages(ctx context.Context, productID int64) ([]*domain.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, created_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []*domain.ProductImage
	for rows.Next() {
		var image domain.ProductImage
		err := rows.Scan(&image.ID, &image.ProductID, &image.ImageURL, &image.IsPrimary, &image.CreatedAt)
		if err != nil {
			return nil, err
		}
		images = append(images, &image)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return images, nil
}

func (r *productRepository) SetImageAsPrimary(ctx context.Context, productID int64, imageID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Set all images for this product as non-primary
	_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = false WHERE product_id = $1", productID)
	if err != nil {
		return err
	}

	// Set the specified image as primary
	_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = true WHERE id = $1 AND product_id = $2", imageID, productID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *productRepository) UpdateProductPrimaryImage(ctx context.Context, productID int64, imageID *int64) error {
	var err error
	if imageID == nil {
		_, err = r.db.ExecContext(ctx, "UPDATE products SET primary_image_id = NULL WHERE id = $1", productID)
	} else {
		_, err = r.db.ExecContext(ctx, "UPDATE products SET primary_image_id = $1 WHERE id = $2", *imageID, productID)
	}
	return err
}
