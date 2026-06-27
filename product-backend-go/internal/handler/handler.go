package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/product-backend/internal/model"
	"github.com/example/product-backend/internal/repository"
	"github.com/example/product-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.Search)
	r.Get("/{id}", h.FindByID)
	r.Put("/{id}", h.Update)
	r.Patch("/{id}/stock", h.UpdateStock)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, model.ApiResponse{Data: resp})
}

func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	params := repository.SearchParams{
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

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *ProductHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	resp, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req model.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	var req model.StockUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.UpdateStock(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: nil})
}

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/{id}", h.FindByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, model.ApiResponse{Data: resp})
}

func (h *CategoryHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 0 {
		page = 0
	}
	size := 10
	if s := r.URL.Query().Get("size"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val > 0 {
			size = val
		}
	}

	resp, err := h.service.FindAll(r.Context(), page, size)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *CategoryHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	resp, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var req model.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: nil})
}

type BrandHandler struct {
	service *service.BrandService
}

func NewBrandHandler(service *service.BrandService) *BrandHandler {
	return &BrandHandler{service: service}
}

func (h *BrandHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/{id}", h.FindByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *BrandHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.BrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Create(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, model.ApiResponse{Data: resp})
}

func (h *BrandHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 0 {
		page = 0
	}
	size := 10
	if s := r.URL.Query().Get("size"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val > 0 {
			size = val
		}
	}

	resp, err := h.service.FindAll(r.Context(), page, size)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *BrandHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid brand ID")
		return
	}

	resp, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *BrandHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid brand ID")
		return
	}

	var req model.BrandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: resp})
}

func (h *BrandHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid brand ID")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, model.ApiResponse{Data: nil})
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
