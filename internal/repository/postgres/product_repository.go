package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
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
		INSERT INTO products (name, slug, description, price, stock_quantity, sub_category_id, created_at, updated_at, is_deleted)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query,
		product.Name,
		product.Slug,
		product.Description,
		product.Price,
		product.StockQuantity,
		product.SubCategoryID,
		product.CreatedAt,
		product.UpdatedAt, false).Scan(&product.ID)

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
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE slug = $1 AND is_deleted = false)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if slug exists, %v:", err)
	}
	return exists, err
}

func (r *productRepository) NameExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE LOWER(name) = LOWER($1) AND is_deleted = false)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	if err != nil {
		log.Printf("error while checking if product name exists: %v", err)
	}
	return exists, err
}

/*
GetByID:
- Get product details from products table
- id, name, slug, description, price, stock_quantity, sub_category_id, created_at, updated_at, deleted_at, is_deleted
*/
func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `SELECT id, name, slug, description, price, stock_quantity, sub_category_id, created_at, updated_at, deleted_at, is_deleted
              FROM products WHERE id = $1 AND is_deleted = false`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.Price,
		&product.StockQuantity,
		&product.SubCategoryID,
		&product.CreatedAt,
		&product.UpdatedAt,
		&product.DeletedAt,
		&product.IsDeleted)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrProductNotFound
		}
		log.Printf("Database error while retrieving product using product ID: %v", err)
		return nil, err
	}

	return &product, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `UPDATE products 
			SET name = $1, slug = $2, description = $3, price = $4, 
              stock_quantity = $5, sub_category_id = $6, updated_at = $7
              WHERE id = $8 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Slug,
		product.Description,
		product.Price,
		product.StockQuantity,
		product.SubCategoryID,
		time.Now().UTC(),
		product.ID)

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
	query := `UPDATE products SET deleted_at = $1, is_deleted = true WHERE id = $2 AND is_deleted = false`

	result, err := r.db.ExecContext(ctx, query, time.Now().UTC(), id)
	if err != nil {
		log.Printf("database error while soft deleting : %v", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("database error while calculating rows affected after soft deletion :%v", err)
		return err
	}

	if rowsAffected == 0 {
		return utils.ErrProductNotFound
	}

	return nil
}

func (r *productRepository) NameExistsBeforeUpdate(ctx context.Context, name string, excludeID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE LOWER(name) = LOWER($1) AND id != $2 AND is_deleted = false)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, name, excludeID).Scan(&exists)
	if err != nil {
		log.Printf("Database error while checking whether name exists before update : %v", err)
	}
	return exists, err
}

func (r *productRepository) AddImage(ctx context.Context, productID int64, imageURL string, isPrimary bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("failed to begin transaction : %v", err)
		return err
	}
	defer tx.Rollback()

	// If this is a primary image, update all other images to non-primary
	if isPrimary {
		_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = false WHERE product_id = $1", productID)
		if err != nil {
			log.Printf("failed to update primary image status of rest of the images : %v", err)
			return err
		}
	}
	// Insert the new image
	query := `
		INSERT INTO product_images (product_id, image_url, is_primary)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var imageID int64
	err = tx.QueryRowContext(ctx, query, productID, imageURL, isPrimary).Scan(&imageID)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == "23505" { // Unique violation
			return utils.ErrDuplicateImageURL
		}
		log.Printf("failed to add the image entry in db : %v", err)
		return err
	}

	// If this is a primary image, update the product's primary_image_id
	if isPrimary {
		_, err = tx.ExecContext(ctx, "UPDATE products SET primary_image_id = $1 WHERE id = $2", imageID, productID)
		if err != nil {
			log.Printf("failed to update the product's primary_image_id : %v", err)
			return err
		}
	}

	return tx.Commit()
}

func (r *productRepository) GetImageCount(ctx context.Context, productID int64) (int, error) {
	query := `
			SELECT COUNT(*) FROM product_images WHERE product_id = $1
			`
	var count int
	err := r.db.QueryRowContext(ctx, query, productID).Scan(&count)
	if err != nil {
		log.Printf("failed to get image count :%v", err)
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
		log.Printf("failed to retrieve product image details from db : %v", err)
		return nil, err
	}
	defer rows.Close()

	var images []*domain.ProductImage
	for rows.Next() {
		var image domain.ProductImage
		err := rows.Scan(&image.ID, &image.ProductID, &image.ImageURL, &image.IsPrimary, &image.CreatedAt)
		if err != nil {
			log.Printf("failed to parse image details to struct : %v", err)
			return nil, err
		}
		images = append(images, &image)
	}
	if err = rows.Err(); err != nil {
		log.Printf("database error : %v", err)
		return nil, err
	}
	return images, nil
}

func (r *productRepository) SetImageAsPrimary(ctx context.Context, productID int64, imageID int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("failed to begin transaction : %v", err)
		return err
	}
	defer tx.Rollback()

	// Set all images for this product as non-primary
	_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = false WHERE product_id = $1", productID)
	if err != nil {
		log.Printf("failed to update primary status of rest of the images : %v", err)
		return err
	}

	// Set the specified image as primary
	_, err = tx.ExecContext(ctx, "UPDATE product_images SET is_primary = true WHERE id = $1 AND product_id = $2", imageID, productID)
	if err != nil {
		log.Printf("failed to set the specified image as primary : %v", err)
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

func (r *productRepository) GetPrimaryImage(ctx context.Context, productID int64) (*domain.ProductImage, error) {
	query := `
		SELECT id, product_id, image_url, is_primary, created_at
		FROM product_images
		WHERE product_id = $1 AND is_primary = true
	`
	var img domain.ProductImage
	err := r.db.QueryRowContext(ctx, query, productID).Scan(
		&img.ID, &img.ProductID, &img.ImageURL, &img.IsPrimary, &img.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil // No primary image found
	}
	if err != nil {
		log.Printf("database error : %v", err)
		return nil, err
	}
	return &img, nil
}

func (r *productRepository) UpdateImagePrimary(ctx context.Context, imageID int64, isPrimary bool) error {
	query := `UPDATE product_images SET is_primary = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, isPrimary, imageID)
	return err
}

func (r *productRepository) GetImageByID(ctx context.Context, imageID int64) (*domain.ProductImage, error) {
	query := `SELECT id,product_id,image_url,is_primary FROM product_images WHERE id= $1`
	var image domain.ProductImage
	err := r.db.QueryRowContext(ctx, query, imageID).Scan(&image.ID, &image.ProductID, &image.ImageURL, &image.IsPrimary)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, utils.ErrImageNotFound
		}
		log.Printf("db error while retrieving product image using id :%v", err)
		return nil, err
	}

	return &image, nil
}

func (r *productRepository) DeleteImageByID(ctx context.Context, imageID int64) error {
	// Start a transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback()

	// First, get the product ID and check if the image is primary
	var productID int64
	var isPrimary bool
	query := `SELECT product_id, is_primary FROM product_images WHERE id = $1`
	err = tx.QueryRowContext(ctx, query, imageID).Scan(&productID, &isPrimary)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrImageNotFound
		}
		log.Printf("Error getting image details: %v", err)
		return err
	}

	// If the image is primary, we need to update the products table
	if isPrimary {
		// Set primary_image_id to NULL for the product
		_, err = tx.ExecContext(ctx, `UPDATE products SET primary_image_id = NULL WHERE id = $1`, productID)
		if err != nil {
			log.Printf("Error updating products table: %v", err)
			return err
		}

	}

	// Now we can safely delete the image
	result, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE id = $1`, imageID)
	if err != nil {
		log.Printf("Error deleting image from product_images table: %v", err)
		return err
	}

	// Check if the image was actually deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error while finding number of rows affected: %v", err)
		return err
	}
	if rowsAffected == 0 {
		return utils.ErrImageNotFound
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	return nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]*domain.Product, error) {
	query := `
		SELECT id, name, description, price, stock_quantity, sub_category_id, created_at, updated_at, slug, is_deleted
		FROM products
		WHERE is_deleted = false
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Printf("error while retrieving product details : %v", err)
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.StockQuantity,
			&p.SubCategoryID, &p.CreatedAt, &p.UpdatedAt, &p.Slug, &p.IsDeleted,
		)
		if err != nil {
			log.Printf("error while getting product data : %v", err)
			return nil, err
		}
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		log.Printf("database error : %v", err)
		return nil, err
	}

	return products, nil
}

func (r *productRepository) UpdateStockQuantity(ctx context.Context, productID int64, quantity int) error {
	query := `UPDATE products SET stock_quantity = $1, updated_at = NOW() WHERE id = $2 AND is_deleted = false`
	result, err := r.db.ExecContext(ctx, query, quantity, productID)
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

func (r *productRepository) GetProducts(ctx context.Context, params domain.ProductQueryParams) ([]*domain.Product, int64, error) {
	query := `
        SELECT p.id, p.name, p.slug, p.description, p.price, p.stock_quantity, p.sub_category_id,
               p.created_at, p.updated_at, p.deleted_at, p.primary_image_id, p.is_deleted,
               c.name AS category_name, sc.name AS subcategory_name
        FROM products p
        JOIN sub_categories sc ON p.sub_category_id = sc.id
        JOIN categories c ON sc.parent_category_id = c.id
        WHERE p.is_deleted = false
    `

	countQuery := `
        SELECT COUNT(*)
        FROM products p
        JOIN sub_categories sc ON p.sub_category_id = sc.id
        JOIN categories c ON sc.parent_category_id = c.id
        WHERE p.is_deleted = false
    `

	var conditions []string
	var args []interface{}

	// ... [Keep the existing condition checks]

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
		countQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Add sorting
	if params.Sort != "" {
		query += fmt.Sprintf(" ORDER BY p.%s %s", params.Sort, params.Order)
	} else {
		query += " ORDER BY p.created_at DESC" // Default sorting
	}

	// Count total before applying pagination
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, params.Limit, (params.Page-1)*params.Limit)

	// Execute main query
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		var categoryName, subcategoryName string
		err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.StockQuantity, &p.SubCategoryID,
			&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.PrimaryImageID, &p.IsDeleted,
			&categoryName, &subcategoryName,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, &p)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return products, totalCount, nil
}

func (r *productRepository) GetPublicProductByID(ctx context.Context, id int64) (*domain.PublicProduct, error) {
	query := `
        SELECT p.id, p.name, p.slug, p.description, p.price, p.stock_quantity, 
               p.created_at, p.updated_at, c.name as category_name, sc.name as subcategory_name
        FROM products p
        JOIN sub_categories sc ON p.sub_category_id = sc.id
        JOIN categories c ON sc.parent_category_id = c.id
        WHERE p.id = $1 AND p.is_deleted = false
    `

	var product domain.PublicProduct
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Slug, &product.Description,
		&product.Price, &product.StockQuantity, &product.CreatedAt,
		&product.UpdatedAt, &product.CategoryName, &product.SubcategoryName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrProductNotFound
		}
		return nil, err
	}

	// Fetch product images
	imagesQuery := `
        SELECT image_url
        FROM product_images
        WHERE product_id = $1
        ORDER BY is_primary DESC, created_at ASC
    `
	rows, err := r.db.QueryContext(ctx, imagesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return nil, err
		}
		images = append(images, imageURL)
	}

	product.Images = images

	return &product, nil
}

/*
UpdateStockTx:
- Update stock_quantity in products table
*/
func (r *productRepository) UpdateStockTx(ctx context.Context, tx *sql.Tx, productID int64, quantity int) error {
	query := `
        UPDATE products
        SET stock_quantity = stock_quantity + $1,
            updated_at = NOW()
        WHERE id = $2
    `
	_, err := tx.ExecContext(ctx, query, quantity, productID)
	if err != nil {
		log.Printf("error while updating stock_quantity in products table : %v", err)
		return err
	}
	return nil
}
