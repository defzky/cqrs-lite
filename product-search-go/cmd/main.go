package main

import (
	"log"
	"net/http"

	"github.com/example/product-search/internal/config"
	"github.com/example/product-search/internal/handler"
	"github.com/example/product-search/internal/nats"
	"github.com/example/product-search/internal/repository"
	"github.com/example/product-search/internal/service"
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

	// Initialize repository
	searchRepo := repository.NewProductSearchRepository(db)

	// Initialize service
	searchService := service.NewProductSearchService(searchRepo)

	// Initialize NATS consumer
	natsConsumer := nats.NewConsumer(cfg.NatsURL, searchService)
	go func() {
		if err := natsConsumer.Start(); err != nil {
			log.Printf("NATS consumer error: %v", err)
		}
	}()
	defer natsConsumer.Stop()

	// Initialize handler
	searchHandler := handler.NewSearchHandler(searchService)

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
	r.Route("/api/search", func(r chi.Router) {
		r.Get("/products", searchHandler.Search)
	})

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8081"
	}
	log.Printf("Starting search service on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
