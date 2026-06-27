# CQRS Lite - Product Demo

Stack ringan untuk demo CQRS dengan memory footprint ~300MB (menggunakan Go).

## Arsitektur

```
PostgreSQL (CDC source)
    │
    ├── product-backend:8080 (Write Side - Go)
    │       └── REST API untuk CRUD produk
    │
    ├── Debezium Server (CDC)
    │       └── Capture perubahan dari PostgreSQL
    │       └── Kirim event ke NATS JetStream
    │
    └── NATS JetStream (Event Streaming)
            │
            └── product-search:8081 (Read Side - Go)
                    └── Consume events dari NATS
                    └── Update search index (PostgreSQL FTS)
```

## Tech Stack

| Komponen | Tool | Memory | Port |
|----------|------|--------|------|
| Database | PostgreSQL 17 | ~150MB | 5433 |
| Message Broker | NATS JetStream | ~100MB | 4222, 8222 |
| CDC | Debezium Server 2.7 (optional) | ~200MB | - |
| Backend (Write) | Go 1.22 + Chi | ~30MB | 8080 |
| Search (Read) | Go 1.22 + Chi + pgx | ~30MB | 8081 |
| Frontend | Nginx | ~10MB | 3000 |

**Total: ~300MB** (sisanya untuk OS)

## Keuntungan Migrasi ke Go

| Aspek | Java (Spring Boot) | Go (Chi) |
|-------|-------------------|----------|
| Memory per service | ~200MB | ~30MB |
| Startup time | ~5-10s | <1s |
| Binary size | ~50MB (JAR) | ~15MB |
| Total footprint | ~500MB | ~300MB |
| Garbage collection | G1GC (tunable) | Go GC (automatic) |

## Quick Start

```bash
# Build dan jalankan semua services
make build

# Cek status
make status

# Cek health
make health

# Test API
make test-api
```

## API Endpoints

### Backend (Write Side) - Port 8080

- `GET /api/products` - List semua produk
- `POST /api/products` - Buat produk baru
- `GET /api/products/{id}` - Detail produk
- `PUT /api/products/{id}` - Update produk
- `DELETE /api/products/{id}` - Hapus produk
- `PATCH /api/products/{id}/stock` - Update stok

- `GET /api/categories` - List kategori
- `POST /api/categories` - Buat kategori

- `GET /api/brands` - List brand
- `POST /api/brands` - Buat brand

### Search (Read Side) - Port 8081

- `GET /api/search/products` - Search dengan facets
  - Query params: `keyword`, `categoryId`, `brandId`, `minPrice`, `maxPrice`, `inStock`, `page`, `size`, `sort`

### Frontend - Port 3000

- `GET /` - Web UI untuk demo

## Perbandingan dengan Stack Asli

| Aspek | Stack Asli | Stack Lite (Go) |
|-------|------------|-----------------|
| Kafka | Kafka (1.5GB) | NATS JetStream (100MB) |
| CDC | Debezium Connect (500MB) | Debezium Server (200MB) |
| Search | OpenSearch (1GB) | PostgreSQL FTS (0MB) |
| Backend | Spring Boot (200MB) | Go Chi (30MB) |
| Search Service | Spring Boot (200MB) | Go Chi (30MB) |
| **Total** | **3-4GB** | **~300MB** |

## Konsep yang Tetap Ada

1. **CQRS** - Write dan Read side terpisah
2. **CDC** - Tangkap perubahan database secara real-time
3. **Event Streaming** - Event-driven architecture
4. **Full-Text Search + Facets** - Search yang proper

## Limitasi

- Tidak ada typo tolerance (OpenSearch lebih baik untuk ini)
- Facets count dihitung on-the-fly (bisa lambat untuk data besar)
- Tidak ada Kafka ecosystem tools

## Files

```
cqrs-lite/
├── docker-compose.yml          # Orkestrasi semua services
├── Makefile                    # Helper commands
├── init-db/                    # SQL schema & seed data
│   └── 01-schema.sql
├── product-backend-go/         # Write side (Go)
│   ├── go.mod
│   ├── Dockerfile
│   ├── cmd/main.go
│   └── internal/
│       ├── config/
│       ├── handler/
│       ├── model/
│       ├── repository/
│       └── service/
├── product-search-go/          # Read side (Go)
│   ├── go.mod
│   ├── Dockerfile
│   ├── cmd/main.go
│   └── internal/
│       ├── config/
│       ├── handler/
│       ├── model/
│       ├── nats/
│       ├── repository/
│       └── service/
├── product-backend/            # Write side (Java - legacy)
├── product-search/             # Read side (Java - legacy)
└── frontend/                   # Web UI
    ├── Dockerfile
    ├── nginx.conf
    └── static/
        └── index.html
```

## Development

### Build Go services locally

```bash
# Build binaries
make build-go

# Or test compilation
make test-go
```

### Run with Go locally (without Docker)

```bash
# Set environment variables
export DB_URL="postgres://postgres:postgres123@localhost:5432/product?sslmode=disable"
export NATS_URL="nats://localhost:4222"

# Run backend
cd product-backend-go && go run ./cmd

# Run search (in another terminal)
cd product-search-go && go run ./cmd
```

## Troubleshooting

```bash
# Cek logs
make logs

# Restart semua services
make down && make up

# Clean semua data
make clean
```

## Libraries Used (Go)

- **Chi** - Lightweight HTTP router
- **pgx** - PostgreSQL driver and connection pool
- **nats.go** - NATS client library
- **google/uuid** - UUID handling
