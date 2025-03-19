# Build stage
FROM golang:1.23-alpine AS builder

# Install required dependencies
RUN apk add --no-cache git make npm build-base

# Set working directory
WORKDIR /app

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Install Goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copy go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Install npm dependencies
RUN npm install

# Generate templ files
RUN templ generate

# Build the application
RUN make build

# Final stage
FROM alpine:3.19

# Install SQLite and other runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite

# Set working directory
WORKDIR /app

# Create directory for SQLite database
RUN mkdir -p /app/data

# Copy the binary from builder
COPY --from=builder /app/bin/app_prod .

# Copy the goose binary from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/

# Copy the public directory for static assets
COPY --from=builder /app/public ./public

# Copy migrations directory from the correct path
COPY --from=builder /app/app/db/migrations ./migrations

# Copy .env file if it exists
COPY --from=builder /app/.env* ./


# Expose the port the app runs on
ENV HTTP_LISTEN_ADDR=:7331
EXPOSE 7331

# Set environment variables for Goose migrations
ENV DB_DRIVER=sqlite3
ENV DB_NAME=/app/data/app.db
ENV MIGRATION_DIR=migrations

# Run migrations and start the application
CMD ["sh", "-c", "goose -dir=\"$MIGRATION_DIR\" \"$DB_DRIVER\" \"$DB_NAME\" up && ./app_prod"]