# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install build dependencies and Atlas
RUN apk add --no-cache git curl && \
    curl -sSf https://atlasgo.sh | sh

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/go-radio cmd/server/main.go

# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app/client

# Copy package files
COPY client/package.json client/yarn.lock ./

# Install dependencies
RUN yarn install --frozen-lockfile

# Copy source code
COPY client/ .

# Build the application
RUN yarn build

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies including PostgreSQL client
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Copy the binary from builder
COPY --from=builder /app/bin/go-radio /app/go-radio

# Copy Atlas binary from builder
COPY --from=builder /usr/local/bin/atlas /usr/local/bin/atlas

# Copy necessary files
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder /app/atlas.hcl /app/atlas.hcl

# Copy built frontend static files
COPY --from=frontend-builder /app/client/dist /app/static

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership of the app directory
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Run the application
CMD ["/app/go-radio"] 