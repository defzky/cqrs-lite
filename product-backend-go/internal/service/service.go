package service

import (
	"context"
	"errors"

	"github.com/example/product-backend/internal/model"
	"github.com/example/product-backend/internal/repository"
	"github.com/google/uuid"
)

type ProductService struct {
	productRepo  *repository.ProductRepository
	categoryRepo *repository.CategoryRepository
	brandRepo    *repository.BrandRepository
}

func NewProductService(productRepo *repository.ProductRepository, categoryRepo *repository.CategoryRepository, brandRepo *repository.BrandRepository) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		brandRepo:    brandRepo,
	}
}

func (s *ProductService) Create(ctx context.Context, req *model.ProductRequest) (*model.ProductResponse, error) {
	// Validate SKU unique
	exists, err := s.productRepo.ExistsBySKU(ctx, req.SKU)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("SKU already exists: " + req.SKU)
	}

	// Validate price
	if req.Price < 0 {
		return nil, errors.New("price must be greater than or equal to 0")
	}

	// Validate stock
	if req.Stock < 0 {
		return nil, errors.New("stock must be greater than or equal to 0")
	}

	// Validate category exists
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, errors.New("category not found: " + req.CategoryID.String())
	}

	// Validate brand exists
	brand, err := s.brandRepo.FindByID(ctx, req.BrandID)
	if err != nil {
		return nil, errors.New("brand not found: " + req.BrandID.String())
	}

	// Create product
	product := &model.Product{
		SKU:         req.SKU,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageURL,
		CategoryID:  req.CategoryID,
		BrandID:     req.BrandID,
	}

	if product.ImageURL == "" {
		product.ImageURL = "https://dummyimage.com/600x400/cccccc/000000&text=No+Image"
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	product.Category = category
	product.Brand = brand

	return s.toResponse(product), nil
}

func (s *ProductService) Search(ctx context.Context, params repository.SearchParams) (*model.PaginatedResponse, error) {
	products, total, err := s.productRepo.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	responses := make([]model.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = *s.toResponse(&p)
	}

	totalPages := int(total) / params.Size
	if int(total)%params.Size > 0 {
		totalPages++
	}

	return &model.PaginatedResponse{
		Data: responses,
		Pagination: model.Pagination{
			Page:       params.Page,
			Size:       params.Size,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *ProductService) FindByID(ctx context.Context, id uuid.UUID) (*model.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found: " + id.String())
	}
	return s.toResponse(product), nil
}

func (s *ProductService) Update(ctx context.Context, id uuid.UUID, req *model.ProductRequest) (*model.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found: " + id.String())
	}

	// Validate SKU unique (if changed)
	if product.SKU != req.SKU {
		exists, err := s.productRepo.ExistsBySKU(ctx, req.SKU)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("SKU already exists: " + req.SKU)
		}
	}

	// Validate price
	if req.Price < 0 {
		return nil, errors.New("price must be greater than or equal to 0")
	}

	// Validate category exists
	category, err := s.categoryRepo.FindByID(ctx, req.CategoryID)
	if err != nil {
		return nil, errors.New("category not found: " + req.CategoryID.String())
	}

	// Validate brand exists
	brand, err := s.brandRepo.FindByID(ctx, req.BrandID)
	if err != nil {
		return nil, errors.New("brand not found: " + req.BrandID.String())
	}

	product.SKU = req.SKU
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.ImageURL = req.ImageURL
	product.CategoryID = req.CategoryID
	product.BrandID = req.BrandID

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	product.Category = category
	product.Brand = brand

	return s.toResponse(product), nil
}

func (s *ProductService) UpdateStock(ctx context.Context, id uuid.UUID, req *model.StockUpdateRequest) (*model.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("product not found: " + id.String())
	}

	newStock := product.Stock
	if req.Type == model.StockIncrease {
		newStock += req.Quantity
	} else {
		newStock -= req.Quantity
		if newStock < 0 {
			return nil, errors.New("stock cannot be negative")
		}
	}

	product.Stock = newStock
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return &model.ProductResponse{
		ID:        product.ID,
		SKU:       product.SKU,
		Stock:     product.Stock,
		UpdatedAt: product.UpdatedAt,
	}, nil
}

func (s *ProductService) Delete(ctx context.Context, id uuid.UUID) error {
	exists, err := s.productRepo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("product not found: " + id.String())
	}
	return s.productRepo.Delete(ctx, id)
}

func (s *ProductService) toResponse(p *model.Product) *model.ProductResponse {
	return &model.ProductResponse{
		ID:          p.ID,
		SKU:         p.SKU,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageURL:    p.ImageURL,
		Category: &model.CategoryRef{
			ID:   p.Category.ID,
			Name: p.Category.Name,
		},
		Brand: &model.BrandRef{
			ID:   p.Brand.ID,
			Name: p.Brand.Name,
		},
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

type CategoryService struct {
	repo *repository.CategoryRepository
}

func NewCategoryService(repo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, req *model.CategoryRequest) (*model.CategoryResponse, error) {
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("category name already exists: " + req.Name)
	}

	category := &model.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	return s.toResponse(category), nil
}

func (s *CategoryService) FindAll(ctx context.Context, page, size int) (*model.PaginatedResponse, error) {
	categories, total, err := s.repo.FindAll(ctx, page, size)
	if err != nil {
		return nil, err
	}

	responses := make([]model.CategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = *s.toResponse(&c)
	}

	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	return &model.PaginatedResponse{
		Data: responses,
		Pagination: model.Pagination{
			Page:       page,
			Size:       size,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *CategoryService) FindByID(ctx context.Context, id uuid.UUID) (*model.CategoryResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("category not found: " + id.String())
	}
	return s.toResponse(category), nil
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, req *model.CategoryRequest) (*model.CategoryResponse, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("category not found: " + id.String())
	}

	exists, err := s.repo.ExistsByNameExcludeID(ctx, req.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("category name already exists: " + req.Name)
	}

	category.Name = req.Name
	category.Description = req.Description

	if err := s.repo.Update(ctx, category); err != nil {
		return nil, err
	}

	return s.toResponse(category), nil
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("category not found: " + id.String())
	}
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) toResponse(c *model.Category) *model.CategoryResponse {
	return &model.CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

type BrandService struct {
	repo *repository.BrandRepository
}

func NewBrandService(repo *repository.BrandRepository) *BrandService {
	return &BrandService{repo: repo}
}

func (s *BrandService) Create(ctx context.Context, req *model.BrandRequest) (*model.BrandResponse, error) {
	exists, err := s.repo.ExistsByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("brand name already exists: " + req.Name)
	}

	brand := &model.Brand{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, brand); err != nil {
		return nil, err
	}

	return s.toResponse(brand), nil
}

func (s *BrandService) FindAll(ctx context.Context, page, size int) (*model.PaginatedResponse, error) {
	brands, total, err := s.repo.FindAll(ctx, page, size)
	if err != nil {
		return nil, err
	}

	responses := make([]model.BrandResponse, len(brands))
	for i, b := range brands {
		responses[i] = *s.toResponse(&b)
	}

	totalPages := int(total) / size
	if int(total)%size > 0 {
		totalPages++
	}

	return &model.PaginatedResponse{
		Data: responses,
		Pagination: model.Pagination{
			Page:       page,
			Size:       size,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *BrandService) FindByID(ctx context.Context, id uuid.UUID) (*model.BrandResponse, error) {
	brand, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("brand not found: " + id.String())
	}
	return s.toResponse(brand), nil
}

func (s *BrandService) Update(ctx context.Context, id uuid.UUID, req *model.BrandRequest) (*model.BrandResponse, error) {
	brand, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("brand not found: " + id.String())
	}

	exists, err := s.repo.ExistsByNameExcludeID(ctx, req.Name, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("brand name already exists: " + req.Name)
	}

	brand.Name = req.Name
	brand.Description = req.Description

	if err := s.repo.Update(ctx, brand); err != nil {
		return nil, err
	}

	return s.toResponse(brand), nil
}

func (s *BrandService) Delete(ctx context.Context, id uuid.UUID) error {
	exists, err := s.repo.ExistsByID(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("brand not found: " + id.String())
	}
	return s.repo.Delete(ctx, id)
}

func (s *BrandService) toResponse(b *model.Brand) *model.BrandResponse {
	return &model.BrandResponse{
		ID:          b.ID,
		Name:        b.Name,
		Description: b.Description,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}
