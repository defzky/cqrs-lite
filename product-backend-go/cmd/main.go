package main

import (
	"log"
	"net/http"

	"github.com/example/product-backend/internal/config"
	"github.com/example/product-backend/internal/handler"
	"github.com/example/product-backend/internal/repository"
	"github.com/example/product-backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Load config
	cfg := config.Load()

	// Connect to database
	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	brandRepo := repository.NewBrandRepository(db)

	// Initialize services
	productService := service.NewProductService(productRepo, categoryRepo, brandRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	brandService := service.NewBrandService(brandRepo)

	// Initialize handlers
	productHandler := handler.NewProductHandler(productService)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	brandHandler := handler.NewBrandHandler(brandService)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/actuator/health", handler.HealthCheck)

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Mount("/products", productHandler.Routes())
		r.Mount("/categories", categoryHandler.Routes())
		r.Mount("/brands", brandHandler.Routes())
	})

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
