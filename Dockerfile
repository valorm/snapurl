# Stage 1: Builder with CGO enabled (required for go-sqlite3)
FROM golang:1.24.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor/ ./vendor/

# Copy source code
COPY . .

# Build with CGO enabled using vendored dependencies (no network download needed)
RUN go build -mod=vendor -o /snapurl ./cmd/server/main.go

# Stage 2: Final image with Alpine (must keep libc for sqlite3)
FROM alpine:latest
# Install runtime dependency for SQLite
RUN apk add --no-cache sqlite-libs
WORKDIR /
COPY --from=builder /snapurl /snapurl
# Don't copy config - it will be mounted as volume in docker-compose
# Create directories that might be needed
RUN mkdir -p /config /data
EXPOSE 8080
ENTRYPOINT ["/snapurl"]