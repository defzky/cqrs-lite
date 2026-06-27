package service

import (
	"context"
	"time"

	"github.com/example/product-search/internal/model"
	"github.com/example/product-search/internal/repository"
	"github.com/google/uuid"
)

type ProductSearchService struct {
	repo *repository.ProductSearchRepository
}

func NewProductSearchService(repo *repository.ProductSearchRepository) *ProductSearchService {
	return &ProductSearchService{repo: repo}
}

func (s *ProductSearchService) Search(ctx context.Context, params model.SearchParams) (*model.SearchResponse, error) {
	startTime := time.Now()

	// Search products
	products, total, err := s.repo.SearchWithFilters(ctx, params)
	if err != nil {
		return nil, err
	}

	// Calculate facets
	categoryFacets, err := s.getCategoryFacets(ctx, params.Keyword, params.CategoryIDs)
	if err != nil {
		return nil, err
	}

	brandFacets, err := s.getBrandFacets(ctx, params.Keyword, params.BrandIDs)
	if err != nil {
		return nil, err
	}

	// Build response
	responses := make([]model.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = model.ProductResponse{
			ID:          p.ID,
			SKU:         p.SKU,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			Stock:       p.Stock,
			ImageURL:    p.ImageURL,
			Category: &model.CategoryRef{
				ID:   p.CategoryID,
				Name: p.CategoryName,
			},
			Brand: &model.BrandRef{
				ID:   p.BrandID,
				Name: p.BrandName,
			},
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	totalPages := int(total) / params.Size
	if int(total)%params.Size > 0 {
		totalPages++
	}

	return &model.SearchResponse{
		Data: responses,
		Paging: model.PagingInfo{
			Page:        params.Page,
			Size:        params.Size,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
		Facets: model.Facets{
			Categories:   categoryFacets,
			Brands:       brandFacets,
			PriceRanges:  s.calculatePriceRanges(),
			Availability: s.calculateAvailability(),
		},
		Metadata: model.Metadata{
			ProcessTimeMs: time.Since(startTime).Milliseconds(),
		},
	}, nil
}

func (s *ProductSearchService) getCategoryFacets(ctx context.Context, keyword string, selectedIDs []uuid.UUID) ([]model.CategoryFacet, error) {
	results, err := s.repo.GetCategoryFacets(ctx, keyword, selectedIDs)
	if err != nil {
		return nil, err
	}

	facets := make([]model.CategoryFacet, len(results))
	for i, r := range results {
		selected := false
		for _, id := range selectedIDs {
			if id == r.ID {
				selected = true
				break
			}
		}
		facets[i] = model.CategoryFacet{
			ID:       r.ID,
			Name:     r.Name,
			Count:    r.Count,
			Selected: selected,
		}
	}
	return facets, nil
}

func (s *ProductSearchService) getBrandFacets(ctx context.Context, keyword string, selectedIDs []uuid.UUID) ([]model.BrandFacet, error) {
	results, err := s.repo.GetBrandFacets(ctx, keyword, selectedIDs)
	if err != nil {
		return nil, err
	}

	facets := make([]model.BrandFacet, len(results))
	for i, r := range results {
		selected := false
		for _, id := range selectedIDs {
			if id == r.ID {
				selected = true
				break
			}
		}
		facets[i] = model.BrandFacet{
			ID:       r.ID,
			Name:     r.Name,
			Count:    r.Count,
			Selected: selected,
		}
	}
	return facets, nil
}

func (s *ProductSearchService) calculatePriceRanges() []model.PriceRangeFacet {
	// Static price ranges for demo
	return []model.PriceRangeFacet{
		{Min: 0, Max: 100000, Label: "< 100.000", Count: 0, Selected: false},
		{Min: 100000, Max: 500000, Label: "100.000 - 500.000", Count: 0, Selected: false},
		{Min: 500000, Max: 0, Label: "> 500.000", Count: 0, Selected: false},
	}
}

func (s *ProductSearchService) calculateAvailability() []model.AvailabilityFacet {
	return []model.AvailabilityFacet{
		{Value: "IN_STOCK", Count: 0, Selected: false},
		{Value: "OUT_OF_STOCK", Count: 0, Selected: false},
	}
}

// IndexProduct - in PostgreSQL FTS, indexing happens automatically via triggers
func (s *ProductSearchService) IndexProduct(ctx context.Context, productID string) error {
	// PostgreSQL FTS index is updated automatically via database triggers
	// This method exists for logging/visibility
	return nil
}

// RemoveProduct - remove from search index
func (s *ProductSearchService) RemoveProduct(ctx context.Context, productID string) error {
	// In CQRS, read model can differ from write model
	// For now, just log this
	return nil
}
