package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID  `json:"id"`
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       float64    `json:"price"`
	Stock       int        `json:"stock"`
	ImageURL    string     `json:"imageUrl"`
	CategoryID  uuid.UUID  `json:"categoryId"`
	BrandID     uuid.UUID  `json:"brandId"`
	Category    *Category  `json:"category,omitempty"`
	Brand       *Brand     `json:"brand,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type Category struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Brand struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DTOs
type ProductRequest struct {
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `json:"imageUrl"`
	CategoryID  uuid.UUID `json:"categoryId"`
	BrandID     uuid.UUID `json:"brandId"`
}

type ProductResponse struct {
	ID          uuid.UUID      `json:"id"`
	SKU         string         `json:"sku"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	ImageURL    string         `json:"imageUrl"`
	Category    *CategoryRef   `json:"category"`
	Brand       *BrandRef      `json:"brand"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type CategoryRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type BrandRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type CategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type BrandRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BrandResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type StockUpdateRequest struct {
	Quantity int              `json:"quantity"`
	Type     StockUpdateType  `json:"type"`
}

type StockUpdateType string

const (
	StockIncrease StockUpdateType = "INCREASE"
	StockDecrease StockUpdateType = "DECREASE"
)

type ApiResponse struct {
	Data interface{} `json:"data"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page        int   `json:"page"`
	Size        int   `json:"size"`
	TotalItems  int64 `json:"totalItems"`
	TotalPages  int   `json:"totalPages"`
}
