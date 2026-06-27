# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and curl for healthcheck
RUN apk add --no-cache ca-certificates curl

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /app/server /app/server

RUN chown -R appuser:appgroup /app
USER appuser

EXPOSE 8081

ENV GOMAXPROCS=2
ENV GOMEMLIMIT=150MiB

ENTRYPOINT ["/app/server"]
