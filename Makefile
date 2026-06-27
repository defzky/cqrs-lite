# CQRS Lite - Makefile
# Usage: make <target>

.PHONY: up down build clean logs status

COMPOSE ?= docker compose

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
	@echo "=== Backend ==="
	@curl -fsS http://localhost:8080/actuator/health 2>/dev/null | grep -o '"status":"[^"]*"' || echo "Not ready"
	@echo ""
	@echo "=== Search ==="
	@curl -fsS http://localhost:8081/actuator/health 2>/dev/null | grep -o '"status":"[^"]*"' || echo "Not ready"

# Test APIs
test-api:
	@echo "=== List Categories ==="
	@curl -s http://localhost:8080/api/categories 2>/dev/null || echo "Backend not running"
	@echo "\n\n=== List Products ==="
	@curl -s http://localhost:8080/api/products 2>/dev/null || echo "Backend not running"
	@echo "\n\n=== Search Products ==="
	@curl -s "http://localhost:8081/api/search/products" 2>/dev/null || echo "Search not running"
