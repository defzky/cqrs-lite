package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/example/product-search/internal/model"
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

type ProductSearchRepository struct {
	db *DB
}

func NewProductSearchRepository(db *DB) *ProductSearchRepository {
	return &ProductSearchRepository{db: db}
}

func (r *ProductSearchRepository) SearchWithFilters(ctx context.Context, params model.SearchParams) ([]model.ProductDocument, int64, error) {
	// Build dynamic query with PostgreSQL FTS
	baseQuery := `
		FROM products p
		JOIN categories c ON p.category_id = c.id
		JOIN brands b ON p.brand_id = b.id
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	// Full-text search using PostgreSQL tsvector
	if params.Keyword != "" {
		baseQuery += fmt.Sprintf(` AND to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
			@@ plainto_tsquery('english', $%d)`, argNum)
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

	// Get paginated results with ranking
	orderCol := "p.name"
	if params.OrderBy != "" {
		orderCol = params.OrderBy
	}
	orderDir := "ASC"
	if params.OrderDir == "DESC" {
		orderDir = "DESC"
	}

	offset := params.Page * params.Size
	dataQuery := fmt.Sprintf(`
		SELECT p.id, p.sku, p.name, p.description, p.price, p.stock, p.image_url,
			   p.category_id, p.brand_id, p.created_at, p.updated_at,
			   c.name as category_name, b.name as brand_name
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

	products := []model.ProductDocument{}
	for rows.Next() {
		p := model.ProductDocument{}
		err := rows.Scan(
			&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.Stock,
			&p.ImageURL, &p.CategoryID, &p.BrandID, &p.CreatedAt, &p.UpdatedAt,
			&p.CategoryName, &p.BrandName,
		)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, p)
	}

	return products, total, rows.Err()
}

type CategoryFacetResult struct {
	ID    uuid.UUID
	Name  string
	Count int64
}

func (r *ProductSearchRepository) GetCategoryFacets(ctx context.Context, keyword string, selectedCategoryIDs []uuid.UUID) ([]CategoryFacetResult, error) {
	query := `
		SELECT p.category_id, c.name as category_name, COUNT(*) as count
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE ($1 = '' OR to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
			@@ plainto_tsquery('english', $1))
		AND ($2::uuid[] IS NULL OR p.category_id != ALL($2::uuid[]))
		GROUP BY p.category_id, c.name
		ORDER BY count DESC
	`

	rows, err := r.db.Query(ctx, query, keyword, selectedCategoryIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facets := []CategoryFacetResult{}
	for rows.Next() {
		f := CategoryFacetResult{}
		err := rows.Scan(&f.ID, &f.Name, &f.Count)
		if err != nil {
			return nil, err
		}
		facets = append(facets, f)
	}

	return facets, rows.Err()
}

type BrandFacetResult struct {
	ID    uuid.UUID
	Name  string
	Count int64
}

func (r *ProductSearchRepository) GetBrandFacets(ctx context.Context, keyword string, selectedBrandIDs []uuid.UUID) ([]BrandFacetResult, error) {
	query := `
		SELECT p.brand_id, b.name as brand_name, COUNT(*) as count
		FROM products p
		JOIN brands b ON p.brand_id = b.id
		WHERE ($1 = '' OR to_tsvector('english', coalesce(p.name, '') || ' ' || coalesce(p.description, '')) 
			@@ plainto_tsquery('english', $1))
		AND ($2::uuid[] IS NULL OR p.brand_id != ALL($2::uuid[]))
		GROUP BY p.brand_id, b.name
		ORDER BY count DESC
	`

	rows, err := r.db.Query(ctx, query, keyword, selectedBrandIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	facets := []BrandFacetResult{}
	for rows.Next() {
		f := BrandFacetResult{}
		err := rows.Scan(&f.ID, &f.Name, &f.Count)
		if err != nil {
			return nil, err
		}
		facets = append(facets, f)
	}

	return facets, rows.Err()
}

func (r *ProductSearchRepository) FindProductByID(ctx context.Context, id uuid.UUID) (*model.ProductDocument, error) {
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.price, p.stock, p.image_url,
			   p.category_id, p.brand_id, p.created_at, p.updated_at,
			   c.name as category_name, b.name as brand_name
		FROM products p
		JOIN categories c ON p.category_id = c.id
		JOIN brands b ON p.brand_id = b.id
		WHERE p.id = $1
	`

	p := &model.ProductDocument{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.Stock,
		&p.ImageURL, &p.CategoryID, &p.BrandID, &p.CreatedAt, &p.UpdatedAt,
		&p.CategoryName, &p.BrandName,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}
