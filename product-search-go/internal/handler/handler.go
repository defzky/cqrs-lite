package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/product-search/internal/model"
	"github.com/example/product-search/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SearchHandler struct {
	service *service.ProductSearchService
}

func NewSearchHandler(service *service.ProductSearchService) *SearchHandler {
	return &SearchHandler{service: service}
}

func (h *SearchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/products", h.Search)
	return r
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	params := model.SearchParams{
		Page:     0,
		Size:     10,
		OrderBy:  "name",
		OrderDir: "ASC",
	}

	// Parse query params
	if keyword := r.URL.Query().Get("keyword"); keyword != "" {
		params.Keyword = keyword
	}
	if categoryIDs := r.URL.Query()["categoryId"]; len(categoryIDs) > 0 {
		for _, idStr := range categoryIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				params.CategoryIDs = append(params.CategoryIDs, id)
			}
		}
	}
	if brandIDs := r.URL.Query()["brandId"]; len(brandIDs) > 0 {
		for _, idStr := range brandIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				params.BrandIDs = append(params.BrandIDs, id)
			}
		}
	}
	if minPrice := r.URL.Query().Get("minPrice"); minPrice != "" {
		if val, err := strconv.ParseFloat(minPrice, 64); err == nil {
			params.MinPrice = &val
		}
	}
	if maxPrice := r.URL.Query().Get("maxPrice"); maxPrice != "" {
		if val, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			params.MaxPrice = &val
		}
	}
	if inStock := r.URL.Query().Get("inStock"); inStock != "" {
		if val, err := strconv.ParseBool(inStock); err == nil {
			params.InStock = &val
		}
	}
	if page := r.URL.Query().Get("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil && val >= 0 {
			params.Page = val
		}
	}
	if size := r.URL.Query().Get("size"); size != "" {
		if val, err := strconv.Atoi(size); err == nil && val > 0 {
			params.Size = val
		}
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		parts := strings.Split(sort, ",")
		if len(parts) == 2 {
			params.OrderBy = parts[0]
			params.OrderDir = strings.ToUpper(parts[1])
		}
	}

	resp, err := h.service.Search(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Health check
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

// Helpers
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
