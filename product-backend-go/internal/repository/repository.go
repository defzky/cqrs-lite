package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/example/product-backend/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	*pgxpool.Pool
}

func NewDB(databaseURL string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

type ProductRepository struct {
	db *DB
}

func NewProductRepository(db *DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *model.Product) error {
	query := `
		INSERT INTO products (sku, name, description, price, stock, image_url, category_id, brand_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		product.SKU, product.Name, product.Description, product.Price,
		product.Stock, product.ImageURL, product.CategoryID, product.BrandID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.price, p.stock, p.image_url,
			   p.category_id, p.brand_id, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.created_at, c.updated_at,
			   b.id, b.name, b.description, b.created_at, b.updated_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1
	`
	product := &model.Product{}
	product.Category = &model.Category{}
	product.Brand = &model.Brand{}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.SKU, &product.Name, &product.Description, &product.Price, &product.Stock,
		&product.ImageURL, &product.CategoryID, &product.BrandID, &product.CreatedAt, &product.UpdatedAt,
		&product.Category.ID, &product.Category.Name, &product.Category.Description,
		&product.Category.CreatedAt, &product.Category.UpdatedAt,
		&product.Brand.ID, &product.Brand.Name, &product.Brand.Description,
		&product.Brand.CreatedAt, &product.Brand.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (r *ProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1)", sku).Scan(&exists)
	return exists, err
}

func (r *ProductRepository) Update(ctx context.Context, product *model.Product) error {
	query := `
		UPDATE products 
		SET sku = $1, name = $2, description = $3, price = $4, stock = $5,
			image_url = $6, category_id = $7, brand_id = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		product.SKU, product.Name, product.Description, product.Price,
		product.Stock, product.ImageURL, product.CategoryID, product.BrandID,
		product.ID,
	).Scan(&product.UpdatedAt)
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id uuid.UUID, stock int) error {
	query := `UPDATE products SET stock = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`
	return r.db.QueryRow(ctx, query, stock, id).Scan(nil)
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM products WHERE id = $1", id)
	return err
}

func (r *ProductRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

type SearchParams struct {
	Keyword     string
	CategoryIDs []uuid.UUID
	BrandIDs    []uuid.UUID
	MinPrice    *float64
	MaxPrice    *float64
	InStock     *bool
	Page        int
	Size        int
	OrderBy     string
	OrderDir    string
}

func (r *ProductRepository) Search(ctx context.Context, params SearchParams) ([]model.Product, int64, error) {
	// Build dynamic query
	baseQuery := `
		FROM products p
		JOIN categories c ON p.category_id = c.id
		JOIN brands b ON p.brand_id = b.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	// Add filters
	if params.Keyword != "" {
		baseQuery += fmt.Sprintf(" AND (LOWER(p.name) LIKE LOWER('%%' || $%d || '%%') OR LOWER(p.description) LIKE LOWER('%%' || $%d || '%%'))", argNum, argNum)
		args = append(args, params.Keyword)
		argNum++
	}

	if len(params.CategoryIDs) > 0 {
		baseQuery += fmt.Sprintf(" AND p.category_id = ANY($%d)", argNum)
		args = append(args, params.CategoryIDs)
		argNum++
	}

	if len(params.BrandIDs) > 0 {
		baseQuery += fmt.Sprintf(" AND p.brand_id = ANY($%d)", argNum)
		args = append(args, params.BrandIDs)
		argNum++
	}

	if params.MinPrice != nil {
		baseQuery += fmt.Sprintf(" AND p.price >= $%d", argNum)
		args = append(args, *params.MinPrice)
		argNum++
	}

	if params.MaxPrice != nil {
		baseQuery += fmt.Sprintf(" AND p.price <= $%d", argNum)
		args = append(args, *params.MaxPrice)
		argNum++
	}

	if params.InStock != nil {
		if *params.InStock {
			baseQuery += " AND p.stock > 0"
		} else {
			baseQuery += " AND p.stock = 0"
		}
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	orderCol := "p.name"
	if params.OrderBy != "" {
		// Add table prefix for known columns to avoid ambiguity
		switch params.OrderBy {
		case "name", "sku", "price", "stock", "created_at", "updated_at":
			orderCol = "p." + params.OrderBy
		default:
			orderCol = params.OrderBy
		}
	}
	orderDir := "ASC"
	if params.OrderDir == "DESC" {
		orderDir = "DESC"
	}

	offset := params.Page * params.Size
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.sku, p.name, p.description, p.price, p.stock, p.image_url,
			   p.category_id, p.brand_id, p.created_at, p.updated_at,
			   c.id, c.name, c.description, c.created_at, c.updated_at,
			   b.id, b.name, b.description, b.created_at, b.updated_at
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, baseQuery, orderCol, orderDir, argNum, argNum+1)
	args = append(args, params.Size, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	products := []model.Product{}
	for rows.Next() {
		p := model.Product{Category: &model.Category{}, Brand: &model.Brand{}}
		err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.ImageURL, &p.CategoryID, &p.BrandID, &p.CreatedAt, &p.UpdatedAt,
			&p.Category.ID, &p.Category.Name, &p.Category.Description,
			&p.Category.CreatedAt, &p.Category.UpdatedAt,
			&p.Brand.ID, &p.Brand.Name, &p.Brand.Description,
			&p.Brand.CreatedAt, &p.Brand.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, rows.Err()
}

type CategoryRepository struct {
	db *DB
}

func NewCategoryRepository(db *DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, category *model.Category) error {
	query := `INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, category.Name, category.Description).
		Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *CategoryRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE id = $1`
	category := &model.Category{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (r *CategoryRepository) FindAll(ctx context.Context, page, size int) ([]model.Category, int64, error) {
	// Count
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Query
	offset := page * size
	query := `SELECT id, name, description, created_at, updated_at FROM categories ORDER BY name LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	categories := []model.Category{}
	for rows.Next() {
		c := model.Category{}
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		categories = append(categories, c)
	}

	return categories, total, rows.Err()
}

func (r *CategoryRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)", name).Scan(&exists)
	return exists, err
}

func (r *CategoryRepository) ExistsByNameExcludeID(ctx context.Context, name string, excludeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1 AND id != $2)", name, excludeID).Scan(&exists)
	return exists, err
}

func (r *CategoryRepository) Update(ctx context.Context, category *model.Category) error {
	query := `UPDATE categories SET name = $1, description = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(ctx, query, category.Name, category.Description, category.ID).Scan(&category.UpdatedAt)
}

func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM categories WHERE id = $1", id)
	return err
}

func (r *CategoryRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", id).Scan(&exists)
	return exists, err
}

type BrandRepository struct {
	db *DB
}

func NewBrandRepository(db *DB) *BrandRepository {
	return &BrandRepository{db: db}
}

func (r *BrandRepository) Create(ctx context.Context, brand *model.Brand) error {
	query := `INSERT INTO brands (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(ctx, query, brand.Name, brand.Description).
		Scan(&brand.ID, &brand.CreatedAt, &brand.UpdatedAt)
}

func (r *BrandRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Brand, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM brands WHERE id = $1`
	brand := &model.Brand{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&brand.ID, &brand.Name, &brand.Description, &brand.CreatedAt, &brand.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return brand, nil
}

func (r *BrandRepository) FindAll(ctx context.Context, page, size int) ([]model.Brand, int64, error) {
	var total int64
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM brands").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := page * size
	query := `SELECT id, name, description, created_at, updated_at FROM brands ORDER BY name LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	brands := []model.Brand{}
	for rows.Next() {
		b := model.Brand{}
		err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		brands = append(brands, b)
	}

	return brands, total, rows.Err()
}

func (r *BrandRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM brands WHERE name = $1)", name).Scan(&exists)
	return exists, err
}

func (r *BrandRepository) ExistsByNameExcludeID(ctx context.Context, name string, excludeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM brands WHERE name = $1 AND id != $2)", name, excludeID).Scan(&exists)
	return exists, err
}

func (r *BrandRepository) Update(ctx context.Context, brand *model.Brand) error {
	query := `UPDATE brands SET name = $1, description = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(ctx, query, brand.Name, brand.Description, brand.ID).Scan(&brand.UpdatedAt)
}

func (r *BrandRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM brands WHERE id = $1", id)
	return err
}

func (r *BrandRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM brands WHERE id = $1)", id).Scan(&exists)
	return exists, err
}
