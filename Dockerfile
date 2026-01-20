# Build stage
FROM docker.io/golang:1.24-alpine AS builder

# Install required dependencies
RUN apk add --no-cache git make npm build-base

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Install npm dependencies
RUN npm install

# Fix ELF image error
RUN npm rebuild esbuild

# In the builder stage, modify your RUN command:
RUN make build

# Final stage
FROM docker.io/alpine:latest

# Install SQLite and other runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite

# Set working directory
WORKDIR /app

# Create directory for SQLite database
RUN mkdir -p /app/data

# Create empty .env file to prevent crash
RUN touch .env

# Copy the binary from builder
COPY --from=builder /app/bin/app_prod .

# Copy the public directory for static assets
COPY --from=builder /app/public ./public

# Expose the port the app runs on
ENV HTTP_LISTEN_ADDR=:7331
EXPOSE 7331

# Set environment variables for Goose migrations
# ENV DB_DRIVER=sqlite3
# ENV DB_NAME=/app/data/app.db
# ENV MIGRATION_DIR=migrations
RUN ls -la /app/public/assets/
# Run migrations and start the application
CMD ["./app_prod"]
