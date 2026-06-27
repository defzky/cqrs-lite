package model

import (
	"time"

	"github.com/google/uuid"
)

type ProductDocument struct {
	ID           uuid.UUID `json:"id"`
	SKU          string    `json:"sku"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	Stock        int       `json:"stock"`
	ImageURL     string    `json:"imageUrl"`
	CategoryID   uuid.UUID `json:"categoryId"`
	BrandID      uuid.UUID `json:"brandId"`
	CategoryName string    `json:"categoryName"`
	BrandName    string    `json:"brandName"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Search Response DTOs
type SearchResponse struct {
	Data     []ProductResponse `json:"data"`
	Paging   PagingInfo        `json:"paging"`
	Facets   Facets            `json:"facets"`
	Metadata Metadata           `json:"metadata"`
}

type ProductResponse struct {
	ID          uuid.UUID    `json:"id"`
	SKU         string       `json:"sku"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Stock       int          `json:"stock"`
	ImageURL    string       `json:"imageUrl"`
	Category    *CategoryRef `json:"category"`
	Brand       *BrandRef    `json:"brand"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}

type CategoryRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type BrandRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type PagingInfo struct {
	Page        int   `json:"page"`
	Size        int   `json:"size"`
	TotalItems  int64 `json:"totalElement"`
	TotalPages  int   `json:"totalPage"`
}

type Facets struct {
	Categories   []CategoryFacet   `json:"categories"`
	Brands       []BrandFacet      `json:"brands"`
	PriceRanges  []PriceRangeFacet `json:"priceRanges"`
	Availability []AvailabilityFacet `json:"availability"`
}

type CategoryFacet struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Count    int64     `json:"count"`
	Selected bool      `json:"selected"`
}

type BrandFacet struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Count    int64     `json:"count"`
	Selected bool      `json:"selected"`
}

type PriceRangeFacet struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Label    string  `json:"label"`
	Count    int64   `json:"count"`
	Selected bool    `json:"selected"`
}

type AvailabilityFacet struct {
	Value    string `json:"value"`
	Count    int64  `json:"count"`
	Selected bool   `json:"selected"`
}

type Metadata struct {
	ProcessTimeMs int64 `json:"processTimeMs"`
}

// Search params
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
