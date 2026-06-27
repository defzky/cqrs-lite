# CQRS Lite - Makefile
# Usage: make <target>

.PHONY: up down build clean logs status test build-go build-java

COMPOSE ?= docker compose

# Default: build Go services
up:
	$(COMPOSE) up -d

build:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

clean:
	$(COMPOSE) down -v

logs:
	$(COMPOSE) logs -f --tail=100

status:
	@echo "=== Container Status ==="
	@docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep cqrs || echo "No containers running"
	@echo ""
	@echo "=== Memory Usage ==="
	@docker stats --no-stream --format "table {{.Name}}\t{{.MemUsage}}" 2>/dev/null | grep -E "NAME|cqrs" || echo "No containers running"

# Health checks
health:
	@echo "=== PostgreSQL ==="
	@docker exec cqrs-postgres pg_isready -U postgres 2>/dev/null || echo "Not ready"
	@echo ""
	@echo "=== NATS ==="
	@curl -fsS http://localhost:8222/healthz 2>/dev/null && echo "OK" || echo "Not ready"
	@echo ""
	@echo "=== Backend (Go) ==="
	@curl -fsS http://localhost:8080/actuator/health 2>/dev/null || echo "Not ready"
	@echo ""
	@echo "=== Search (Go) ==="
	@curl -fsS http://localhost:8081/actuator/health 2>/dev/null || echo "Not ready"

# Test APIs
test-api:
	@echo "=== List Categories ==="
	@curl -s http://localhost:8080/api/categories 2>/dev/null || echo "Backend not running"
	@echo "\n\n=== List Products ==="
	@curl -s http://localhost:8080/api/products 2>/dev/null || echo "Backend not running"
	@echo "\n\n=== Search Products ==="
	@curl -s "http://localhost:8081/api/search/products" 2>/dev/null || echo "Search not running"

# Build Go services locally (for testing)
build-go:
	@echo "=== Building product-backend-go ==="
	@cd product-backend-go && go build -o bin/server ./cmd
	@echo "=== Building product-search-go ==="
	@cd product-search-go && go build -o bin/server ./cmd
	@echo "Done!"

# Test Go services compile
test-go:
	@echo "=== Testing product-backend-go ==="
	@cd product-backend-go && go build -o /dev/null ./cmd
	@echo "=== Testing product-search-go ==="
	@cd product-search-go && go build -o /dev/null ./cmd
	@echo "All Go services compile successfully!"

# Original Java build (kept for reference)
build-java:
	@echo "=== Building product-backend (Java) ==="
	@cd product-backend && mvn clean package -DskipTests
	@echo "=== Building product-search (Java) ==="
	@cd product-search && mvn clean package -DskipTests
	@echo "Done!"
